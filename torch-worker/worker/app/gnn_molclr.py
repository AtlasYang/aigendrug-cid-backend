import torch
import torch.nn as nn
import torch.nn.functional as F
import pandas as pd
import argparse
import os
from torch.utils.data import random_split
from torch_geometric.loader import DataLoader
from MolCLR.dataset import MolDataset
from MolCLR.ginet_finetune import GINEConv, num_atom_type, num_chirality_tag
from torch_geometric.nn import global_mean_pool, global_max_pool, global_add_pool
from gnn_utils_dm import EarlyStopper


class GINet_MTL(nn.Module):
    def __init__(
        self,
        task=["classification"],
        num_layer=5,
        emb_dim=300,
        feat_dim=512,
        drop_ratio=0,
        pool="mean",
        pred_n_layer=2,
        pred_act="softplus",
        z_cutoff=0,
    ):
        super(GINet_MTL, self).__init__()
        self.num_layer = num_layer
        self.emb_dim = emb_dim
        self.feat_dim = feat_dim
        self.drop_ratio = drop_ratio
        self.task = task
        num_tasks = len(task)
        self.z_cutoff = z_cutoff

        self.x_embedding1 = nn.Embedding(num_atom_type, emb_dim)
        self.x_embedding2 = nn.Embedding(num_chirality_tag, emb_dim)
        nn.init.xavier_uniform_(self.x_embedding1.weight.data)
        nn.init.xavier_uniform_(self.x_embedding2.weight.data)

        self.gnns = nn.ModuleList()
        for layer in range(num_layer):
            self.gnns.append(GINEConv(emb_dim))

        self.batch_norms = nn.ModuleList()
        for layer in range(num_layer):
            self.batch_norms.append(nn.BatchNorm1d(emb_dim))

        if pool == "mean":
            self.pool = global_mean_pool
        elif pool == "max":
            self.pool = global_max_pool
        elif pool == "add":
            self.pool = global_add_pool

        self.feat_lin = nn.Linear(self.emb_dim, self.feat_dim)
        out_dim = 1
        self.pred_n_layer = max(1, pred_n_layer)

        self.pred_heads = nn.ModuleList()
        for _ in range(num_tasks):
            if pred_act == "relu":
                pred_head = [
                    nn.Linear(self.feat_dim, self.feat_dim // 2),
                    nn.ReLU(inplace=True),
                ]
                for _ in range(self.pred_n_layer - 1):
                    pred_head.extend(
                        [
                            nn.Linear(self.feat_dim // 2, self.feat_dim // 2),
                            nn.ReLU(inplace=True),
                        ]
                    )
                pred_head.append(nn.Linear(self.feat_dim // 2, out_dim))
            elif pred_act == "softplus":
                pred_head = [
                    nn.Linear(self.feat_dim, self.feat_dim // 2),
                    nn.Softplus(),
                ]
                for _ in range(self.pred_n_layer - 1):
                    pred_head.extend(
                        [
                            nn.Linear(self.feat_dim // 2, self.feat_dim // 2),
                            nn.Softplus(),
                        ]
                    )
                pred_head.append(nn.Linear(self.feat_dim // 2, out_dim))
            else:
                raise ValueError("Undefined activation function")

            self.pred_heads.append(nn.Sequential(*pred_head))

    def load_my_state_dict(self, state_dict):
        own_state = self.state_dict()
        for name, param in state_dict.items():
            if name not in own_state:
                continue
            if isinstance(param, nn.parameter.Parameter):
                param = param.data
            own_state[name].copy_(param)

    def forward(self, data):
        x = data.x
        edge_index = data.edge_index
        edge_attr = data.edge_attr

        h = self.x_embedding1(x[:, 0]) + self.x_embedding2(x[:, 1])

        for layer in range(self.num_layer):
            h = self.gnns[layer](h, edge_index, edge_attr)
            h = self.batch_norms[layer](h)
            if layer == self.num_layer - 1:
                h = F.dropout(h, self.drop_ratio, training=self.training)
            else:
                h = F.dropout(F.relu(h), self.drop_ratio, training=self.training)

        h = self.pool(h, data.batch)
        h = self.feat_lin(h)

        return h, torch.cat([head(h) for head in self.pred_heads], dim=1)


def load_pre_trained_weights(model, device, checkpoint_file):
    state_dict = torch.load(checkpoint_file, map_location=device, weights_only=True)
    model.load_my_state_dict(state_dict)
    return model


def train(model, trainloader, args, optimizer, criterion_list):
    model.train()
    train_loss = 0
    for batch in trainloader:
        batch = batch.to(args.device)
        label = batch.y

        optimizer.zero_grad()
        _, pred = model(batch)

        total_loss = 0
        for i in range(len(model.task)):
            null_mask = torch.isnan(label[:, i])
            predi = pred[~null_mask, i]
            labeli = label[~null_mask, i]
            loss = criterion_list[i](predi, labeli)
            total_loss += loss

        total_loss.backward()
        optimizer.step()
        train_loss += loss.item()
    return train_loss / len(trainloader)


def eval(model, loader, args, criterion_list, return_output=False):
    model.eval()
    preds = []
    ys = []
    with torch.no_grad():
        for batch in loader:
            batch = batch.to(args.device)
            label = batch.y
            _, pred = model(batch)
            preds.append(pred)
            ys.append(label)
    preds = torch.cat(preds, dim=0)
    ys = torch.cat(ys, dim=0)

    total_loss = 0
    losses = []
    for i in range(len(model.task)):
        null_mask = torch.isnan(ys[:, i])
        predi = preds[~null_mask, i]
        labeli = ys[~null_mask, i]
        loss = criterion_list[i](predi, labeli)
        total_loss += loss
        losses.append(loss.item())

    if return_output:
        return total_loss.item(), losses, preds, ys
    else:
        return total_loss.item(), losses


def predict(model, loader, args):
    model.eval()
    preds = []
    with torch.no_grad():
        for batch in loader:
            _, pred = model(batch.to(args.device))
            preds.append(pred)
    return torch.cat(preds).cpu().numpy().flatten()


def molclr_train_and_predict(train_csv, test_csv):
    args = argparse.Namespace()
    args.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    device = args.device

    dataset = MolDataset(
        pd.read_csv(train_csv), "log_standard_value", "regression", "smiles"
    )
    train_size = int(0.8 * len(dataset))
    val_size = len(dataset) - train_size
    trainset, validset = random_split(dataset, [train_size, val_size])

    trainloader = DataLoader(trainset, batch_size=128, shuffle=True)
    validloader = DataLoader(validset, batch_size=128, shuffle=False)

    model = GINet_MTL(["regression"]).to(device)
    model = load_pre_trained_weights(model, device, "/app/app/MolCLR/pretrained_gin_model.pth")

    layer_list = []
    for name, param in model.named_parameters():
        if "pred_head" in name:
            layer_list.append(name)
    params = list(
        map(
            lambda x: x[1],
            list(filter(lambda kv: kv[0] in layer_list, model.named_parameters())),
        )
    )
    base_params = list(
        map(
            lambda x: x[1],
            list(filter(lambda kv: kv[0] not in layer_list, model.named_parameters())),
        )
    )
    optimizer = torch.optim.Adam(
        [{"params": base_params, "lr": 0.0001}, {"params": params}],
        0.0005,
        weight_decay=1e-6,
    )

    criterion_list = []
    for task in model.task:
        if task == "classification":
            criterion_list.append(nn.BCEWithLogitsLoss())
        elif task == "regression":
            criterion_list.append(nn.MSELoss())

    early_stopper = EarlyStopper(patience=20, printfunc=print, verbose=False, path=None)
    while not early_stopper.early_stop:
        train(model, trainloader, args, optimizer, criterion_list)
        valid_loss, _ = eval(model, validloader, args, criterion_list)
        early_stopper(valid_loss, model)

    full_train_loader = DataLoader(dataset, batch_size=128, shuffle=False)
    test_dataset = MolDataset(
        pd.read_csv(test_csv), "log_standard_value", "regression", "smiles"
    )
    test_loader = DataLoader(test_dataset, batch_size=128, shuffle=False)

    return predict(model, full_train_loader, args), predict(model, test_loader, args)
