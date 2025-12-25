import React from 'react';

interface OsqueryFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  statusFilter: string;
  onStatusChange: (value: string) => void;
}

const OsqueryFilter: React.FC<OsqueryFilterProps> = ({
  searchTerm,
  onSearchChange,
  statusFilter,
  onStatusChange,
}) => {
  return (
    <div className="flex gap-4 mb-4">
      <input
        type="text"
        placeholder="Search processes..."
        value={searchTerm}
        onChange={(e) => onSearchChange(e.target.value)}
        className="flex-1 px-4 py-2 border rounded"
      />
      <select
        value={statusFilter}
        onChange={(e) => onStatusChange(e.target.value)}
        className="px-4 py-2 border rounded"
      >
        <option value="all">All Status</option>
        <option value="running">Running</option>
        <option value="stopped">Stopped</option>
      </select>
    </div>
  );
};

export default OsqueryFilter;
