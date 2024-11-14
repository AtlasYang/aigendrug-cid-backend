CREATE TABLE job (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    target_protein_name VARCHAR(255) NOT NULL
);

CREATE TABLE experiment (
    id SERIAL PRIMARY KEY,
    type INT NOT NULL, -- 0: with measured value, 1: without measured value
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ligand_smiles TEXT NOT NULL,
    ligand_ranking_in_job INT,
    predicted_value DECIMAL(15, 6),
    measured_value DECIMAL(15, 6),
    training_status INT NOT NULL, -- 0: not trained, 1: training, 2: trained, 3: failed
    job_id INT NOT NULL,

    FOREIGN KEY (job_id) REFERENCES Job(id) ON DELETE CASCADE
);