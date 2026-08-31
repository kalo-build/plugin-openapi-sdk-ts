export interface Company {
  created_at: string;
  id: string;
  name: string;
  updated_at: string;
}

export interface CompanyCreate {
  id?: string;
  name: string;
}

export interface CompanyUpdate {
  name?: string;
}
