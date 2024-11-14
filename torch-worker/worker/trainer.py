import torch
import torch.nn as nn
import torch.optim as optim
import numpy as np
from torch.optim.lr_scheduler import CosineAnnealingLR, CosineAnnealingWarmRestarts
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error, ndcg_score
from scipy.stats import rankdata
from utils import get_complex_embedding, get_complex_embedding_molformer, get_complex_embedding_chemberta_molformer

import pandas as pd
import matplotlib.pyplot as plt

device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')

def train_regressor_model(regressor_model, X_train, y_train, X_test, y_test, epochs=100):
    criterion = nn.MSELoss()
    optimizer = optim.AdamW(regressor_model.parameters(), lr=1e-3, weight_decay=1e-5)
    scheduler = CosineAnnealingWarmRestarts(optimizer, T_0=50, T_mult=1)
    
    best_loss = float('inf')
    patience, early_stop_count = 10, 0
    
    for epoch in range(epochs):
        regressor_model.train()
        optimizer.zero_grad()

        predicted_train = regressor_model(X_train).squeeze()
        loss = criterion(predicted_train, y_train)
        loss.backward()
        optimizer.step()
        scheduler.step()

        regressor_model.eval()
        with torch.no_grad():
            predicted_test = regressor_model(X_test).squeeze()
            test_loss = criterion(predicted_test, y_test)

        rescaled_mse = mean_squared_error(y_test.cpu().numpy(), predicted_test.cpu().numpy())

        print(f"Epoch {epoch+1}/{epochs}, Train Loss: {loss.item():.4f}, Test MSE: {rescaled_mse:.4f}")

        if rescaled_mse < best_loss:
            best_loss = rescaled_mse
            early_stop_count = 0
        else:
            early_stop_count += 1

        if early_stop_count >= patience:
            break

def predict_with_regressor_model(regressor_model, smiles_list):
    regressor_model.eval()
    X = get_complex_embedding_molformer(smiles_list)
    X = torch.tensor(X, dtype=torch.float32).to(device)
    with torch.no_grad():
        predicted = regressor_model(X).squeeze()
    return predicted.cpu().numpy()
