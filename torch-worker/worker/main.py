from flask import Flask, request, jsonify
import torch
import torch.nn as nn
import torch.nn.functional as F
import pandas as pd
import json
from app.app import train_and_predict

app = Flask(__name__)

@app.route('/')
def index():
    return "Model server is running"

@app.route('/process', methods=['POST'])
def load():
    data = request.json
    train_csv_path = data.get('train_csv_path')
    test_csv_path = data.get('test_csv_path')

    
    app.logger.info(f"Training data path: {train_csv_path}")
    app.logger.info(f"Test data path: {test_csv_path}")

    train_df = pd.read_csv(train_csv_path)
    test_df = pd.read_csv(test_csv_path)
    
    with open('/app/csv/log', 'a') as f:
        f.write(f"Request time: {pd.Timestamp.now()}\n")
        f.write(f"Training data path: {train_csv_path}\n")
        f.write(f"Test data path: {test_csv_path}\n")
        f.write(f"Training data shape: {train_df.shape}\n")
        f.write(f"Test data shape: {test_df.shape}\n")
        f.write(f"Training data head:\n{train_df.head()}\n")
    
    df = train_and_predict(train_csv_path, test_csv_path)
    # overwrite test_csv_path with the predicted values
    df.to_csv(test_csv_path, index=False)

    with open('/app/csv/log', 'a') as f:
        f.write(f"Predicted data head:\n{df.head()}\n")
        f.write("\n")

    return jsonify({"message": "Model trained and predicted"})


if __name__ == '__main__':
    app.logger.info("Model server is running")
    app.run(host='0.0.0.0', port=5000)
