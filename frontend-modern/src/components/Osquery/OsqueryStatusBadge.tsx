import React from 'react';

interface OsqueryStatusBadgeProps {
  status: string;
}

const OsqueryStatusBadge: React.FC<OsqueryStatusBadgeProps> = ({ status }) => {
  const getStatusColor = () => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'running':
        return 'bg-green-100 text-green-800';
      case 'inactive':
      case 'stopped':
        return 'bg-gray-100 text-gray-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-blue-100 text-blue-800';
    }
  };

  return (
    <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor()}`}>
      {status}
    </span>
  );
};

export default OsqueryStatusBadge;
