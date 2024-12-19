from numpy import ndarray, array
from rdkit import Chem
from rdkit.Chem import rdFingerprintGenerator
from transformers import AutoModel, AutoTokenizer
import torch

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

chemberta_model_name = "seyonec/ChemBERTa-zinc-base-v1"
chemberta_model = AutoModel.from_pretrained(chemberta_model_name).to(device)
chemberta_tokenizer = AutoTokenizer.from_pretrained(chemberta_model_name)

molformer_model_name = "ibm/MoLFormer-XL-both-10pct"
molformer_model = AutoModel.from_pretrained(
    molformer_model_name, deterministic_eval=True, trust_remote_code=True
).to(device)
molformer_tokenizer = AutoTokenizer.from_pretrained(
    molformer_model_name, trust_remote_code=True
)


def get_fingerprint_from_smiles(smiles: str, fpSize: int = 2048) -> ndarray:
    mfpgen = rdFingerprintGenerator.GetMorganGenerator(fpSize=fpSize)
    mol = Chem.MolFromSmiles(smiles)
    if mol is None:
        with open("/app/csv/log.txt", "a") as f:
            f.write(f"Failed to parse SMILES: {smiles}\n")
        return None
    fingerprint = mfpgen.GetFingerprintAsNumPy(mol)
    return fingerprint


def get_chemberta_embedding(smiles_list) -> ndarray:
    embeddings = []
    for smiles in smiles_list:
        inputs = chemberta_tokenizer(
            smiles, return_tensors="pt", padding=True, truncation=True
        ).to(device)
        with torch.no_grad():
            outputs = chemberta_model(**inputs)
            cls_embedding = outputs.last_hidden_state[:, 0, :].squeeze(0).cpu()
        embeddings.append(cls_embedding.numpy())
    return array(embeddings)


def get_molformer_embedding(smiles_list) -> ndarray:
    embeddings = []
    for smiles in smiles_list:
        inputs = molformer_tokenizer(smiles, return_tensors="pt", padding=True).to(
            device
        )
        with torch.no_grad():
            outputs = molformer_model(**inputs)
            cls_embedding = outputs.pooler_output.squeeze(0).cpu()
        embeddings.append(cls_embedding.numpy())
    return array(embeddings)


def get_complex_embedding(smiles_list) -> ndarray:
    # concat chemberta embedding and fingerprint
    chemberta_embeddings = get_chemberta_embedding(smiles_list)
    fingerprints = [get_fingerprint_from_smiles(smiles) for smiles in smiles_list]
    return array(
        [
            torch.cat(
                (torch.tensor(chemberta_embedding), torch.tensor(fingerprint))
            ).numpy()
            for chemberta_embedding, fingerprint in zip(
                chemberta_embeddings, fingerprints
            )
        ]
    )


def get_complex_embedding_molformer(smiles_list) -> ndarray:
    # concat molformer embedding and fingerprint
    molformer_embeddings = get_molformer_embedding(smiles_list)
    fingerprints = [get_fingerprint_from_smiles(smiles) for smiles in smiles_list]
    return array(
        [
            torch.cat(
                (torch.tensor(molformer_embedding), torch.tensor(fingerprint))
            ).numpy()
            for molformer_embedding, fingerprint in zip(
                molformer_embeddings, fingerprints
            )
        ]
    )


def get_complex_embedding_chemberta_molformer(smiles_list) -> ndarray:
    # concat chemberta embedding, molformer embedding, and fingerprint
    chemberta_embeddings = get_chemberta_embedding(smiles_list)
    molformer_embeddings = get_molformer_embedding(smiles_list)
    fingerprints = [get_fingerprint_from_smiles(smiles) for smiles in smiles_list]
    return array(
        [
            torch.cat(
                (
                    torch.tensor(chemberta_embedding),
                    torch.tensor(molformer_embedding),
                    torch.tensor(fingerprint),
                )
            ).numpy()
            for chemberta_embedding, molformer_embedding, fingerprint in zip(
                chemberta_embeddings, molformer_embeddings, fingerprints
            )
        ]
    )
