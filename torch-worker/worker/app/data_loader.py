"""
Data Loader for trainer and inference functions

todo: load weight from minio client
"""

import pandas as pd
from numpy import ndarray


def load_raw_csv(file_path) -> tuple[pd.Series, ndarray]:
    data = pd.read_csv(file_path)
    smiles = data["smiles"]
    labels = None
    if "log_standard_value" in data.columns:
        labels = data["log_standard_value"].values
    return smiles, labels
