import numpy as np

import torch
import torch.nn.functional as F
from torch.utils.data.sampler import SubsetRandomSampler

from torch_geometric.data import Data, Dataset, DataLoader

import rdkit
from rdkit import Chem
from rdkit.Chem.rdchem import BondType as BT
from rdkit.Chem.Scaffolds.MurckoScaffold import MurckoScaffoldSmiles
from rdkit import RDLogger

RDLogger.DisableLog("rdApp.*")

ATOM_LIST = list(range(1, 119))
CHIRALITY_LIST = [
    Chem.rdchem.ChiralType.CHI_UNSPECIFIED,
    Chem.rdchem.ChiralType.CHI_TETRAHEDRAL_CW,
    Chem.rdchem.ChiralType.CHI_TETRAHEDRAL_CCW,
    Chem.rdchem.ChiralType.CHI_OTHER,
]
BOND_LIST = [BT.SINGLE, BT.DOUBLE, BT.TRIPLE, BT.AROMATIC]
BONDDIR_LIST = [
    Chem.rdchem.BondDir.NONE,
    Chem.rdchem.BondDir.ENDUPRIGHT,
    Chem.rdchem.BondDir.ENDDOWNRIGHT,
]


def read_smiles(data_path, target):
    import pandas as pd

    df = pd.read_csv(data_path)
    smiles = df["SMILES"].values
    label = df[target].values
    print("SMILES in dataframe:", len(smiles))
    return smiles, label


def read_smiles_df(df, target_col, smiles_col="SMILES"):
    """
    df: dataframe with smiles and label
    target_col: target column name
    smiles_col: column name for smiles
    """

    smiles = df[smiles_col].values
    label = None
    if "log_standard_value" in df.columns:
        label = df[target_col].values
    return smiles, label


class MolDataset(Dataset):
    def __init__(self, df, target, task, smiles_col="SMILES"):
        super(Dataset, self).__init__()
        self.smiles_data, self.labels = read_smiles_df(df, target, smiles_col)
        self.task = task
        self.mol_data = []
        for i in range(len(self.smiles_data)):
            smiles = self.smiles_data[i]
            edge_index, edge_attr, x = self.smiles2data(smiles)
            if task == "classification":
                y = torch.tensor(self.labels[i], dtype=torch.float).view(1, -1)
            elif task == "regression" and self.labels is not None:
                y = torch.tensor(self.labels[i], dtype=torch.float).view(1, -1)
                data = Data(x=x, edge_index=edge_index, edge_attr=edge_attr, y=y)
            else:
                data = Data(x=x, edge_index=edge_index, edge_attr=edge_attr)
            self.mol_data.append(data)

    def smiles2data(self, smiles):
        mol = Chem.MolFromSmiles(smiles)
        mol = Chem.AddHs(mol)

        N = mol.GetNumAtoms()
        M = mol.GetNumBonds()

        type_idx = []
        chirality_idx = []
        atomic_number = []
        for atom in mol.GetAtoms():
            type_idx.append(ATOM_LIST.index(atom.GetAtomicNum()))
            chirality_idx.append(CHIRALITY_LIST.index(atom.GetChiralTag()))
            atomic_number.append(atom.GetAtomicNum())

        x1 = torch.tensor(type_idx, dtype=torch.long).view(-1, 1)
        x2 = torch.tensor(chirality_idx, dtype=torch.long).view(-1, 1)
        x = torch.cat([x1, x2], dim=-1)

        row, col, edge_feat = [], [], []
        for bond in mol.GetBonds():
            start, end = bond.GetBeginAtomIdx(), bond.GetEndAtomIdx()
            row += [start, end]
            col += [end, start]
            edge_feat.append(
                [
                    BOND_LIST.index(bond.GetBondType()),
                    BONDDIR_LIST.index(bond.GetBondDir()),
                ]
            )
            edge_feat.append(
                [
                    BOND_LIST.index(bond.GetBondType()),
                    BONDDIR_LIST.index(bond.GetBondDir()),
                ]
            )

        edge_index = torch.tensor([row, col], dtype=torch.long)
        edge_attr = torch.tensor(np.array(edge_feat), dtype=torch.long)
        return edge_index, edge_attr, x

    def __getitem__(self, index):
        return self.mol_data[index]

    def __len__(self):
        return len(self.smiles_data)

    # geometric problem: Can't instantiate abstract class Dataset with abstract methods get, len
    def get():
        pass

    def len():
        pass
