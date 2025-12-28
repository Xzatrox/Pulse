import React from 'react';
import { OsqueryReport } from '../../api/osquery';

interface OsqueryHostSummaryTableProps {
  reports: Record<string, OsqueryReport>;
}

const OsqueryHostSummaryTable: React.FC<OsqueryHostSummaryTableProps> = ({ reports }) => {
  const calculateTotalMemory = (processes: OsqueryReport['processes']) => {
    const total = processes.reduce((sum, p) => {
      const mem = parseInt(p.memory_bytes || '0');
      return sum + (isNaN(mem) ? 0 : mem);
    }, 0);
    if (total < 1024 * 1024) return `${(total / 1024).toFixed(1)} KB`;
    if (total < 1024 * 1024 * 1024) return `${(total / (1024 * 1024)).toFixed(1)} MB`;
    return `${(total / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  };

  return (
    <div className="mb-6">
      <h2 className="text-xl font-semibold mb-4">Host Summary</h2>
      <table className="w-full border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border p-2 text-left">Agent ID</th>
            <th className="border p-2 text-left">Processes</th>
            <th className="border p-2 text-left">Services</th>
            <th className="border p-2 text-left">Total Memory</th>
            <th className="border p-2 text-left">Last Update</th>
          </tr>
        </thead>
        <tbody>
          {Object.entries(reports).map(([agentId, report]) => (
            <tr key={agentId} className="hover:bg-gray-50">
              <td className="border p-2">{agentId}</td>
              <td className="border p-2">{report.processes.length}</td>
              <td className="border p-2">{report.services.length}</td>
              <td className="border p-2">{calculateTotalMemory(report.processes)}</td>
              <td className="border p-2">{new Date(report.timestamp).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default OsqueryHostSummaryTable;
