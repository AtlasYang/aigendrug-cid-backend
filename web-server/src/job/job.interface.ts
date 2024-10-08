export interface Job {
  id: number;
  name: string;
  target_protein_name: string;
  target_protein_file_url: string;
}

export interface JobCreateDto {
  name: string;
  target_protein_name: string;
  target_protein_file_url: string;
}

export const JobSchema = {
  type: 'object',
  properties: {
    id: { type: 'number' },
    name: { type: 'string' },
    target_protein_name: { type: 'string' },
    target_protein_file_url: { type: 'string' },
  },
  required: ['id', 'name', 'target_protein_name', 'target_protein_file_url'],
};

export const JobCreateSchema = {
  type: 'object',
  properties: {
    name: { type: 'string' },
    target_protein_name: { type: 'string' },
    target_protein_file_url: { type: 'string' },
  },
  required: ['name', 'target_protein_name', 'target_protein_file_url'],
};
