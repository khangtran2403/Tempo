import React, { useEffect, useState } from 'react';
import { auditAPI } from '../../api/audit';

type AuditLog = {
  id: string;
  user_id: string;
  resource_id: string;
  action: string;
  ip: string;
  user_agent: string;
  created_at: string;
  metadata?: Record<string, any>;
};

export default function AuditLogsPage() {
  const [items, setItems] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const data = await auditAPI.getAll();
        if (mounted) setItems(data.logs || []);
      } catch (e: any) {
        setError(e?.message || 'Failed to load audit logs');
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => { mounted = false; };
  }, []);

  if (loading) return <div className="p-6">Loading...</div>;
  if (error) return <div className="p-6 text-red-600">{error}</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">Audit Logs</h1>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Resource</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP</th>
              <th className="px-4 py-2"></th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {items.map((log) => (
              <tr key={log.id}>
                <td className="px-4 py-2">{new Date(log.created_at).toLocaleString()}</td>
                <td className="px-4 py-2">{log.action}</td>
                <td className="px-4 py-2 font-mono text-xs">{log.resource_id}</td>
                <td className="px-4 py-2">{log.ip}</td>
                <td className="px-4 py-2 text-right">
                  {/* future: details modal */}
                </td>
              </tr>
            ))}
            {items.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-gray-500">No audit logs</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
