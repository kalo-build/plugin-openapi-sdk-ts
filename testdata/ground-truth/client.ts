import type { Company, CompanyCreate, CompanyUpdate } from "./types";

export class ApiError extends Error {
  statusCode: number;
  body: string;

  constructor(statusCode: number, body: string) {
    super(`API error ${statusCode}: ${body}`);
    this.name = "ApiError";
    this.statusCode = statusCode;
    this.body = body;
  }
}

export class Client {
  baseUrl: string;
  private fetchFn: typeof fetch;

  constructor(baseUrl: string, fetchFn: typeof fetch = fetch.bind(globalThis)) {
    this.baseUrl = baseUrl;
    this.fetchFn = fetchFn;
  }

  async listCompany(): Promise<Company[]> {
    let url = `${this.baseUrl}/api/companies`;
    const resp = await this.fetchFn(url, {
      method: "GET",
      headers: { "Accept": "application/json" },
    });
    if (!resp.ok) {
      const body = await resp.text();
      throw new ApiError(resp.status, body);
    }
    return resp.json() as Promise<Company[]>;
  }

  async createCompany(body: CompanyCreate): Promise<Company> {
    let url = `${this.baseUrl}/api/companies`;
    const resp = await this.fetchFn(url, {
      method: "POST",
      headers: { "Content-Type": "application/json", "Accept": "application/json" },
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      const body = await resp.text();
      throw new ApiError(resp.status, body);
    }
    return resp.json() as Promise<Company>;
  }

  async getCompany(id: string): Promise<Company> {
    let url = `${this.baseUrl}/api/companies/${encodeURIComponent(String(id))}`;
    const resp = await this.fetchFn(url, {
      method: "GET",
      headers: { "Accept": "application/json" },
    });
    if (!resp.ok) {
      const body = await resp.text();
      throw new ApiError(resp.status, body);
    }
    return resp.json() as Promise<Company>;
  }

  async updateCompany(id: string, body: CompanyUpdate): Promise<Company> {
    let url = `${this.baseUrl}/api/companies/${encodeURIComponent(String(id))}`;
    const resp = await this.fetchFn(url, {
      method: "PUT",
      headers: { "Content-Type": "application/json", "Accept": "application/json" },
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      const body = await resp.text();
      throw new ApiError(resp.status, body);
    }
    return resp.json() as Promise<Company>;
  }

  async deleteCompany(id: string): Promise<void> {
    let url = `${this.baseUrl}/api/companies/${encodeURIComponent(String(id))}`;
    const resp = await this.fetchFn(url, {
      method: "DELETE",
      headers: { "Accept": "application/json" },
    });
    if (!resp.ok) {
      const body = await resp.text();
      throw new ApiError(resp.status, body);
    }
  }
}
