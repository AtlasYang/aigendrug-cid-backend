import torch
import torch.nn as nn
import os

class SimpleFFModel(nn.Module):
    def __init__(self):
        super(SimpleFFModel, self).__init__()
        self.fc = nn.Linear(10, 1)

model_weights_path = './weights/'
if not os.path.exists(model_weights_path):
    os.makedirs(model_weights_path)
base_weight_file = os.path.join(model_weights_path, 'base_weight.pth')

model = SimpleFFModel()
torch.save(model.state_dict(), base_weight_file)

print(f"Base weight saved at {base_weight_file}")
