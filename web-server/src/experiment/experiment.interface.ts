export interface Experiment {
  id: number;
  type: number; // 0: with measured value, 1: without measured value
  name: string;
  created_at: string;
  edited_at: string;
  ligand_smiles: string;
  predicted_value: number;
  measured_value: number;
  training_status: number; // 0: not trained, 1: training, 2: trained, 3: failed
  job_id: number;
}

export interface ExperimentCreateDto {
  type: number;
  name: string;
  ligand_smiles: string;
  measured_value: number;
  job_id: number;
}

export interface ExperimentBatchCreateDto {
  type: number;
  name: string;
  ligand_smiles_list: string[];
  measured_value_list: number[];
  job_id: number;
}

export const ExperimentSchema = {
  type: 'object',
  properties: {
    id: { type: 'number' },
    type: { type: 'number' },
    name: { type: 'string' },
    created_at: { type: 'string' },
    edited_at: { type: 'string' },
    ligand_smiles: { type: 'string' },
    predicted_value: { type: 'number' },
    measured_value: { type: 'number' },
    training_status: { type: 'number' },
    job_id: { type: 'number' },
  },
  required: [
    'id',
    'type',
    'name',
    'created_at',
    'edited_at',
    'ligand_smiles',
    'predicted_value',
    'measured_value',
    'training_status',
    'job_id',
  ],
};

export const ExperimentCreateSchema = {
  type: 'object',
  properties: {
    type: { type: 'number' },
    name: { type: 'string' },
    ligand_smiles: { type: 'string' },
    measured_value: { type: 'number' },
    job_id: { type: 'number' },
  },
  required: ['type', 'name', 'ligand_smiles', 'measured_value', 'job_id'],
};

export const ExperimentBatchCreateSchema = {
  type: 'object',
  properties: {
    type: { type: 'number' },
    name: { type: 'string' },
    ligand_smiles_list: { type: 'array', items: { type: 'string' } },
    measured_value_list: { type: 'array', items: { type: 'number' } },
    job_id: { type: 'number' },
  },
  required: [
    'type',
    'name',
    'ligand_smiles_list',
    'measured_value_list',
    'job_id',
  ],
};
