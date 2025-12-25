import React from 'react';
import { OsqueryReport } from '../../api/osquery';
import OsqueryStatusBadge from './OsqueryStatusBadge';

interface OsqueryUnifiedTableProps {
  reports: Record<string, OsqueryReport>;
  searchTerm: string;
  statusFilter: string;
}

const OsqueryUnifiedTable: React.FC<OsqueryUnifiedTableProps> = ({
  reports,
  searchTerm,
  statusFilter,
}) => {
  const allProcesses = Object.entries(reports).flatMap(([agentId, report]) =>
    report.processes.map((p) => ({ ...p, agentId }))
  );

  const filtered = allProcesses.filter((p) => {
    const matchesSearch = p.name.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesSearch;
  });

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Running Processes</h2>
      <table className="w-full border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border p-2 text-left">PID</th>
            <th className="border p-2 text-left">Name</th>
            <th className="border p-2 text-left">Path</th>
            <th className="border p-2 text-left">Log Files</th>
            <th className="border p-2 text-left">Agent</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((process, idx) => (
            <tr key={idx} className="hover:bg-gray-50">
              <td className="border p-2">{process.pid}</td>
              <td className="border p-2 font-medium">{process.name}</td>
              <td className="border p-2 text-sm text-gray-600">{process.path}</td>
              <td className="border p-2 text-sm">
                {process.log_files.length > 0 ? (
                  <ul className="list-disc list-inside">
                    {process.log_files.slice(0, 3).map((log, i) => (
                      <li key={i}>{log}</li>
                    ))}
                    {process.log_files.length > 3 && (
                      <li>+{process.log_files.length - 3} more</li>
                    )}
                  </ul>
                ) : (
                  <span className="text-gray-400">None</span>
                )}
              </td>
              <td className="border p-2">{process.agentId}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default OsqueryUnifiedTable;
