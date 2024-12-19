import torch
import torch.nn as nn
import torch.optim as optim
from torch.optim.lr_scheduler import CosineAnnealingWarmRestarts
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error
from ai_model import ThreeLayerRegressor
from data_loader import load_raw_csv
from ai_model_utils import get_complex_embedding_molformer, get_complex_embedding


class Config:
    def __init__(self, train_csv, test_csv):
        self.TRAIN_CSV = train_csv
        self.TEST_CSV = test_csv
        self.WEIGHTS_ROOT = "ai_model_weights"
        self.DEVICE = "cuda" if torch.cuda.is_available() else "cpu"

        self.TRAIN_CONFIG = {
            "test_size": 0.2,
            "random_state": 42,
            "epochs": 100,
            "learning_rate": 1e-3,
            "weight_decay": 1e-5,
            "patience": 10,
            "scheduler_T0": 50,
            "scheduler_T_mult": 1,
        }

        self.MODEL_CONFIG = {
            "input_dim": 768 + 2048,
            "hidden_1_dim": 512,
            "hidden_2_dim": 256,
            "hidden_3_dim": 128,
            "dropout_prob": 0.3,
        }


class DataModule:
    def __init__(self, config):
        self.config = config

    def get_embeddings(self, smiles, embedding_type):
        if embedding_type == "molformer":
            return get_complex_embedding_molformer(smiles)
        return get_complex_embedding(smiles)

    def prepare_data(self, embedding_type):
        smiles, labels = load_raw_csv(self.config.TRAIN_CSV)
        X = self.get_embeddings(smiles, embedding_type)

        X_train, X_test, y_train, y_test = train_test_split(
            X,
            labels,
            test_size=self.config.TRAIN_CONFIG["test_size"],
            random_state=self.config.TRAIN_CONFIG["random_state"],
        )

        return {
            "X_train": torch.tensor(X_train, dtype=torch.float32).to(
                self.config.DEVICE
            ),
            "X_test": torch.tensor(X_test, dtype=torch.float32).to(self.config.DEVICE),
            "y_train": torch.tensor(y_train, dtype=torch.float32).to(
                self.config.DEVICE
            ),
            "y_test": torch.tensor(y_test, dtype=torch.float32).to(self.config.DEVICE),
        }


class ModelModule:
    def __init__(self, config):
        self.config = config
        self.model = None
        self.criterion = nn.MSELoss()

    def initialize_model(self):
        self.model = ThreeLayerRegressor(**self.config.MODEL_CONFIG).to(
            self.config.DEVICE
        )
        return self.model

    def configure_optimizers(self):
        optimizer = optim.AdamW(
            self.model.parameters(),
            lr=self.config.TRAIN_CONFIG["learning_rate"],
            weight_decay=self.config.TRAIN_CONFIG["weight_decay"],
        )
        scheduler = CosineAnnealingWarmRestarts(
            optimizer,
            T_0=self.config.TRAIN_CONFIG["scheduler_T0"],
            T_mult=self.config.TRAIN_CONFIG["scheduler_T_mult"],
        )
        return optimizer, scheduler

    def train_step(self, batch):
        self.model.train()
        X_train, y_train = batch["X_train"], batch["y_train"]

        predicted = self.model(X_train).squeeze()
        loss = self.criterion(predicted, y_train)

        return loss

    def validation_step(self, batch):
        self.model.eval()
        X_test, y_test = batch["X_test"], batch["y_test"]

        with torch.no_grad():
            predicted = self.model(X_test).squeeze()
            mse = mean_squared_error(y_test.cpu().numpy(), predicted.cpu().numpy())

        return mse

    def predict(self, smiles_list, embedding_type):
        self.model.eval()
        data_module = DataModule(self.config)
        X = data_module.get_embeddings(smiles_list, embedding_type)
        X = torch.tensor(X, dtype=torch.float32).to(self.config.DEVICE)

        with torch.no_grad():
            predicted = self.model(X).squeeze()
        return predicted.cpu().numpy()


class AIModelTrainer:
    def __init__(self, train_csv, test_csv):
        self.config = Config(train_csv, test_csv)
        self.data_module = DataModule(self.config)
        self.model_module = ModelModule(self.config)

    def train(self, embedding_type):
        data = self.data_module.prepare_data(embedding_type)

        self.model_module.initialize_model()
        optimizer, scheduler = self.model_module.configure_optimizers()

        best_loss = float("inf")
        early_stop_count = 0

        for epoch in range(self.config.TRAIN_CONFIG["epochs"]):
            loss = self.model_module.train_step(data)
            optimizer.zero_grad()
            loss.backward()
            optimizer.step()
            scheduler.step()

            val_mse = self.model_module.validation_step(data)

            if val_mse < best_loss:
                best_loss = val_mse
                early_stop_count = 0
            else:
                early_stop_count += 1

            if early_stop_count >= self.config.TRAIN_CONFIG["patience"]:
                break

    def predict(self, embedding_type):
        train_smiles, train_labels = load_raw_csv(self.config.TRAIN_CSV)
        test_smiles, _ = load_raw_csv(self.config.TEST_CSV)
        return (
            self.model_module.predict(train_smiles, embedding_type),
            self.model_module.predict(test_smiles, embedding_type),
        )
