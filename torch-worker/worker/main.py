from flask import Flask, request, jsonify
import torch
import torch.nn as nn
import torch.nn.functional as F
import json
from utils import get_complex_embedding_molformer

app = Flask(__name__)
model = None
model_loaded = False

class ThreeLayerRegressor(nn.Module):
    def __init__(self, input_dim=2816, hidden_1_dim=512, hidden_2_dim=256, hidden_3_dim=128, dropout_prob=0.0):
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
            nn.Sigmoid()
        )

    def forward(self, x):
        return self.model(x)

def load_model(weight_path):
    global model, model_loaded
    model = ThreeLayerRegressor()
    model.load_state_dict(torch.load(weight_path))
    model.eval()
    model_loaded = True

@app.route('/')
def index():
    return "Model server is running with model loaded: {}".format(model_loaded)

@app.route('/load', methods=['POST'])
def load():
    print("Model loading")
    data = request.json
    weight_path: str = data.get('weight_path')
    weight_path = "/app/weights/" + weight_path.split("/")[-1]
    load_model(weight_path)
    return jsonify({"status": "model loaded"})

@app.route('/inference', methods=['POST'])
def inference():
    print("Inference")
    if not model_loaded:
        return jsonify({"error": "model not loaded"}), 400

    data: str= request.json.get('protein_data')
    input_tensor = get_complex_embedding_molformer([data])
    input_tensor = torch.tensor(input_tensor, dtype=torch.float32)
    result = model(input_tensor).item()

    return jsonify({"result": result})

@app.route('/train', methods=['POST'])
def train():
    if not model_loaded:
        return jsonify({"error": "model not loaded"}), 400

    data = request.json
    input_tensor = get_complex_embedding_molformer([data['protein_data']])[0]
    target_value = float(data['target_value'])
    target_tensor = torch.tensor([target_value]).float()
    return jsonify({"status": "training completed"})

if __name__ == '__main__':
    app.logger.info("Model server is running")
    app.run(host='0.0.0.0', port=5000)
