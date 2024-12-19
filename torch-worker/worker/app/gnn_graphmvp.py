import numpy as np
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, random_split
from torch_geometric.loader import DataLoader
from graphmvp.config import args
from graphmvp.models import GNN, GNN_graphpred
from graphmvp.datasets import mol_to_graph_data_obj_simple
from gnn_utils_dm import EarlyStopper
from rdkit.Chem import AllChem
from data_loader import load_raw_csv


class Dataset(Dataset):
    def __init__(self, csv_path):
        self.data = self._process(csv_path)

    def _process(self, csv_path):
        smiles_list, labels = load_raw_csv(csv_path)
        if labels is not None and labels.ndim == 1:
            labels = np.expand_dims(labels, axis=1)

        data_list = []
        for i, smiles in enumerate(smiles_list):
            mol = AllChem.MolFromSmiles(smiles)
            if mol is not None:
                data = mol_to_graph_data_obj_simple(mol)
                data.id = torch.tensor([i])
                if labels is not None:
                    data.y = torch.tensor(labels[i])
                data_list.append(data)

        return data_list

    def __len__(self):
        return len(self.data)

    def __getitem__(self, idx):
        return self.data[idx]


class Model:
    def __init__(self):
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self._build_model()
        self._build_optimizer()
        self._build_criterion()

    def _build_model(self):
        molecule_model = GNN(
            num_layer=args.num_layer,
            emb_dim=args.emb_dim,
            JK=args.JK,
            drop_ratio=args.dropout_ratio,
            gnn_type=args.gnn_type,
        )

        self.model = GNN_graphpred(
            args=args, num_tasks=1, molecule_model=molecule_model
        )
        self.model.from_pretrained("/app/app/graphmvp/pretraining_model.pth", device=self.device)
        self.model.to(self.device)

    def _build_optimizer(self):
        param_group = [
            {"params": self.model.molecule_model.parameters()},
            {
                "params": self.model.graph_pred_linear.parameters(),
                "lr": args.lr * args.lr_scale,
            },
        ]
        self.optimizer = optim.Adam(param_group, lr=args.lr, weight_decay=args.decay)

    def _build_criterion(self):
        self.criterion = nn.MSELoss()

    def train_step(self, data):
        data = data.to(self.device)
        label = data.y.float().unsqueeze(1)

        self.optimizer.zero_grad()
        pred = self.model(data)

        null_mask = torch.isnan(label[:, 0])
        loss = self.criterion(pred[~null_mask, 0], label[~null_mask, 0])

        loss.backward()
        self.optimizer.step()
        return loss.item()

    def eval_step(self, data):
        data = data.to(self.device)
        label = data.y.float().unsqueeze(1)
        pred = self.model(data)

        null_mask = torch.isnan(label[:, 0])
        loss = self.criterion(pred[~null_mask, 0], label[~null_mask, 0])
        return loss.item()

    def train(self, loader):
        self.model.train()
        total_loss = 0
        for batch in loader:
            total_loss += self.train_step(batch)
        return total_loss / len(loader)

    def evaluate(self, loader):
        self.model.eval()
        total_loss = 0
        with torch.no_grad():
            for batch in loader:
                total_loss += self.eval_step(batch)
        return total_loss / len(loader)

    def predict(self, loader):
        self.model.eval()
        preds = []
        with torch.no_grad():
            for batch in loader:
                preds.append(self.model(batch.to(self.device)))
        return torch.cat(preds).cpu().numpy().flatten()


def graphmvp_train_and_predict(train_csv, test_csv):
    dataset = Dataset(train_csv)

    train_size = int(0.8 * len(dataset))
    trainset, validset = random_split(dataset, [train_size, len(dataset) - train_size])

    train_loader = DataLoader(trainset, batch_size=args.batch_size, shuffle=True)
    valid_loader = DataLoader(validset, batch_size=args.batch_size, shuffle=False)

    model = Model()
    early_stopper = EarlyStopper(patience=20, printfunc=print, verbose=False, path=None)

    while not early_stopper.early_stop:
        model.train(train_loader)
        valid_loss = model.evaluate(valid_loader)
        early_stopper(valid_loss, model.model)

    full_train_loader = DataLoader(dataset, batch_size=args.batch_size, shuffle=False)
    test_dataset = Dataset(test_csv)
    test_loader = DataLoader(test_dataset, batch_size=args.batch_size, shuffle=False)

    return model.predict(full_train_loader), model.predict(test_loader)
