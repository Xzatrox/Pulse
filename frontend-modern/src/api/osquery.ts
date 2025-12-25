export interface Process {
  pid: string;
  name: string;
  path: string;
  log_files: string[];
}

export interface Service {
  name: string;
  state: string;
  status: string;
}

export interface OsqueryReport {
  timestamp: string;
  processes: Process[];
  services: Service[];
}

export const OsqueryAPI = {
  getReport: async (agentId: string): Promise<OsqueryReport> => {
    const response = await fetch(`/api/agents/${agentId}/osquery`);
    if (!response.ok) throw new Error('Failed to fetch osquery report');
    return response.json();
  },

  getAllReports: async (): Promise<Record<string, OsqueryReport>> => {
    const response = await fetch('/api/osquery/reports');
    if (!response.ok) throw new Error('Failed to fetch osquery reports');
    return response.json();
  },
};
