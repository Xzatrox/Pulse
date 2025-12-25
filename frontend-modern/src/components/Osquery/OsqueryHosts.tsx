import React, { useEffect, useState } from 'react';
import { OsqueryAPI, OsqueryReport } from '../../api/osquery';
import OsqueryFilter from './OsqueryFilter';
import OsqueryHostSummaryTable from './OsqueryHostSummaryTable';
import OsqueryUnifiedTable from './OsqueryUnifiedTable';

const OsqueryHosts: React.FC = () => {
  const [reports, setReports] = useState<Record<string, OsqueryReport>>({});
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => {
    loadReports();
    const interval = setInterval(loadReports, 30000);
    return () => clearInterval(interval);
  }, []);

  const loadReports = async () => {
    try {
      const data = await OsqueryAPI.getAllReports();
      setReports(data);
    } catch (error) {
      console.error('Failed to load osquery reports:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="p-8 text-center">Loading osquery data...</div>;
  }

  if (Object.keys(reports).length === 0) {
    return <div className="p-8 text-center text-gray-500">No osquery agents reporting</div>;
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">osquery Monitoring</h1>
      <OsqueryFilter
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        statusFilter={statusFilter}
        onStatusChange={setStatusFilter}
      />
      <OsqueryHostSummaryTable reports={reports} />
      <OsqueryUnifiedTable
        reports={reports}
        searchTerm={searchTerm}
        statusFilter={statusFilter}
      />
    </div>
  );
};

export default OsqueryHosts;
