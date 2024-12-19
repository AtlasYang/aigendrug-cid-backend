import pandas as pd
import numpy as np
from ai_model_trainer import AIModelTrainer
from gnn_graphmvp import graphmvp_train_and_predict
from gnn_molclr import molclr_train_and_predict
from sklearn.metrics import mean_squared_error
from autogluon.tabular import TabularDataset, TabularPredictor


def calculate_base_model_mse(df):
    model_columns = ["molformer", "chemberta", "graphmvp", "molclr"]
    for model in model_columns:
        mse = mean_squared_error(df["log_standard_value"], df[model])
        print(f"{model} MSE: {mse:.4f}")


def train_and_predict(
    train_csv,
    test_csv,
    ensemble_quality="medium_quality",
    ensemble_time_limit=1_000_000,
):
    train_df = pd.read_csv(train_csv)
    test_df = pd.read_csv(test_csv)

    # add column 'log_standard_value' if not exists
    if "log_standard_value" not in train_df.columns:
        train_df.insert(
            train_df.columns.get_loc("standard_value") + 1,
            "log_standard_value",
            0.125 * np.log1p(train_df["standard_value"]),
        )
        train_df.to_csv(train_csv, index=False)

    ai_model_trainer = AIModelTrainer(train_csv, test_csv)
    ai_model_trainer.train("molformer")
    train_df["molformer"], test_df["molformer"] = ai_model_trainer.predict("molformer")
    ai_model_trainer.train("chemberta")
    train_df["chemberta"], test_df["chemberta"] = ai_model_trainer.predict("chemberta")
    train_df["graphmvp"], test_df["graphmvp"] = graphmvp_train_and_predict(
        train_csv, test_csv
    )
    train_df["molclr"], test_df["molclr"] = molclr_train_and_predict(
        train_csv, test_csv
    )

    train_data = TabularDataset(train_df.drop(columns=["smiles", "standard_value"]))
    predictor = TabularPredictor(
        label="log_standard_value", path=f"AutogluonModels/ag-default", verbosity=0
    ).fit(train_data, time_limit=ensemble_time_limit, presets=ensemble_quality)

    test_data = TabularDataset(test_df)
    log_pred = predictor.predict(test_data)
    pred = np.expm1(log_pred / 0.125)

    return pd.DataFrame({"smiles": test_data["smiles"], "pred": pred})


if __name__ == "__main__":
    train_csv = "./dataset/protein1_pretrain.csv"
    test_csv = "./dataset/protein1_sampled100.csv"

    train_df = pd.read_csv(train_csv)
    if "log_standard_value" not in train_df.columns:
        train_df.insert(
            train_df.columns.get_loc("standard_value") + 1,
            "log_standard_value",
            0.125 * np.log1p(train_df["standard_value"]),
        )
        train_df.to_csv(train_csv, index=False)

    pred_df = train_and_predict(train_csv, test_csv)
    print(pred_df)
