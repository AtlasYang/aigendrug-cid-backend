
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .

RUN apt-get update && apt-get install -y \
    build-essential \
    python3-dev \
    cmake \
    && apt-get clean

RUN pip install torch

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV PYTHONPATH="/app:/app/app"

EXPOSE 5000

CMD ["python", "main.py"]
