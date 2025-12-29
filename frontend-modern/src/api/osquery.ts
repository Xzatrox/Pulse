import { apiFetchJSON } from '@/utils/apiClient';

export interface Process {
  pid: string;
  name: string;
  path: string;
  log_files: string[];
  log_command?: string;
  memory_bytes?: string;
  status?: string;
}

export interface Service {
  name: string;
  state: string;
  status: string;
  health?: string;
}

export interface OsqueryReport {
  timestamp: string;
  processes: Process[];
  services: Service[];
}

export const OsqueryAPI = {
  getReport: async (agentId: string): Promise<OsqueryReport> => {
    return apiFetchJSON(`/api/agents/${agentId}/osquery`);
  },

  getAllReports: async (): Promise<Record<string, OsqueryReport>> => {
    return apiFetchJSON('/api/osquery/reports');
  },
};
