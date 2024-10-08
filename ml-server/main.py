from flask import Flask
from confluent_kafka import Consumer, Producer
import torch
import torch.nn as nn
import threading
import os
import json
import time

app = Flask(__name__)

kafka_server = os.getenv('KAFKA_SERVER', 'localhost:9092')
kafka_group_id = os.getenv('KAFKA_GROUP_ID', 'aigendrug-ml-server')
model_weights_path = os.getenv('MODEL_WEIGHTS_PATH', './weights/')

kafka_initialized = False
models = []
base_weight_file = os.path.join(model_weights_path, 'base_weight.pth')

# Neural Network Model (Sample)
class SimpleFFModel(nn.Module):
    def __init__(self):
        super(SimpleFFModel, self).__init__()
        self.fc = nn.Linear(10, 1)

    def forward(self, x):
        return self.fc(x)

model_lock = threading.Lock()

def load_model_weights(job_id):
    with model_lock:
        model = SimpleFFModel()
        weight_file = os.path.join(model_weights_path, f'weight_sample_{job_id}.pth')
        if os.path.exists(weight_file):
            model.load_state_dict(torch.load(weight_file))
        else:
            model.load_state_dict(torch.load(base_weight_file))
            torch.save(model.state_dict(), weight_file)
        return model

# Kafka configuration
consumer_conf = {
    'bootstrap.servers': kafka_server,
    'group.id': kafka_group_id,
    'auto.offset.reset': 'earliest'
}

producer_conf = {
    'bootstrap.servers': kafka_server
}

consumer = None
producer = None

def initialize_kafka():
    global consumer, producer
    retries = 30

    while retries > 0:
        try:
            consumer = Consumer(consumer_conf)
            producer = Producer(producer_conf)
            consumer.subscribe(['ModelTrainRequest', 'ModelInferenceRequest'])
            print("Kafka initialized successfully.")
            return
        except Exception as e:
            retries -= 1
            print(f"Error initializing Kafka: {e}. Retries left: {retries}")
            time.sleep(5)

    raise Exception("Failed to initialize Kafka after multiple retries.")

def send_response(topic, response):
    producer.produce(topic, json.dumps(response).encode('utf-8'))
    producer.flush()

# Kafka worker function
def kafka_worker():
    while True:
        try:
            msg = consumer.poll(1.0)
            if msg is None:
                continue

            message = json.loads(msg.value().decode('utf-8'))
            print(f"Received message: {message}")
            
            job_id = message.get('job_id')
            experiment_id = message.get('experiment_id')
            protein_data = message.get('protein_data')
            if job_id is None or protein_data is None or experiment_id is None: 
                raise Exception("Invalid message format")

            protein_data = torch.tensor(protein_data, dtype=torch.float32).unsqueeze(0)

            model = load_model_weights(job_id)
            
            if msg.topic() == 'ModelTrainRequest':
                train_model(model, protein_data, message['target_value'], job_id, experiment_id)
            elif msg.topic() == 'ModelInferenceRequest':
                inference_model(model, protein_data, job_id, experiment_id)
        
        except Exception as e:
            print(f"Error processing message: {e}")

# Model training function
def train_model(model, protein_data, target_value, job_id, experiment_id):
    criterion = nn.MSELoss()
    optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
    target = torch.tensor([target_value], dtype=torch.float32)
    optimizer.zero_grad()
    output = model(protein_data)
    loss = criterion(output, target)
    loss.backward()
    optimizer.step()
    
    response = {
        'success': True,
        'experiment_id': experiment_id,
        'result': output.item()
    }

    time.sleep(20) # Simulate training time

    send_response('ModelTrainResponse', response)

# Model inference function
def inference_model(model, protein_data, job_id, experiment_id):
    with torch.no_grad():
        output = model(protein_data)
    
    response = {
        'success': True,
        'experiment_id': experiment_id,
        'result': output.item()
    }

    time.sleep(5) # Simulate inference time

    send_response('ModelInferenceResponse', response)

def start_kafka_worker():
    initialize_kafka()
    kafka_worker()

def _setup():
    global kafka_initialized
    if not kafka_initialized:
        kafka_initialized = True
        kafka_thread = threading.Thread(target=start_kafka_worker, daemon=True)
        kafka_thread.start()

@app.route('/')
def home():
    return 'Aigendrug ML Server'

@app.route('/models')
def get_models():
    return json.dumps([model.state_dict() for model in models])

_setup()

if __name__ == '__main__':
    if not os.path.exists(model_weights_path):
        os.makedirs(model_weights_path)

    app.run(host='0.0.0.0', port=5000)
