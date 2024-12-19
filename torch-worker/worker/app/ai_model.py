import torch
import torch.nn as nn
import torch.nn.functional as F


class ThreeLayerRegressor(nn.Module):
    def __init__(
        self,
        input_dim=768,
        hidden_1_dim=512,
        hidden_2_dim=256,
        hidden_3_dim=128,
        dropout_prob=0.3,
    ):
        super(ThreeLayerRegressor, self).__init__()

        self.model = nn.Sequential(
            nn.Linear(input_dim, hidden_1_dim),
            nn.BatchNorm1d(hidden_1_dim),
            nn.ReLU(),
            nn.Dropout(dropout_prob),
            nn.Linear(hidden_1_dim, hidden_2_dim),
            nn.BatchNorm1d(hidden_2_dim),
            nn.ReLU(),
            nn.Dropout(dropout_prob),
            nn.Linear(hidden_2_dim, hidden_3_dim),
            nn.BatchNorm1d(hidden_3_dim),
            nn.ReLU(),
            nn.Dropout(dropout_prob),
            nn.Linear(hidden_3_dim, 1),
            nn.Sigmoid(),
        )

    def forward(self, x):
        return self.model(x)


class ListwiseLoss(nn.Module):
    def __init__(self):
        super(ListwiseLoss, self).__init__()

    def forward(self, y_pred, y_true):
        sorted_indices = torch.argsort(y_true, dim=-1, descending=True)

        y_pred_sorted = torch.gather(y_pred, dim=-1, index=sorted_indices)

        loss = -torch.sum(
            y_pred_sorted - torch.logcumsumexp(y_pred_sorted, dim=-1), dim=-1
        )

        return torch.mean(loss)
