export interface Job {
  id: number;
  name: string;
  created_at: string;
  target_protein_name: string;
}

export interface JobCreateDto {
  name: string;
  target_protein_name: string;
}

export interface JobInitialLigands {
  name: string;
  smiles: string;
  std_value: number;
}

export const JobSchema = {
  type: 'object',
  properties: {
    id: { type: 'number' },
    name: { type: 'string' },
    created_at: { type: 'string' },
    target_protein_name: { type: 'string' },
  },
  required: ['id', 'name', 'target_protein_name'],
};

export const JobCreateSchema = {
  type: 'object',
  properties: {
    name: { type: 'string' },
    target_protein_name: { type: 'string' },
  },
  required: ['name', 'target_protein_name'],
};
