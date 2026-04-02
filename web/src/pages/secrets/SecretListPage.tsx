import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import secretsAPI from "../../api/secret";
import { Secret } from "../../types/secrets";

export default function SecretsListPage() {
  const [params] = useSearchParams();
  const workflowId = params.get("workflow_id");

  const [items, setItems] = useState<Secret[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      setItems(await secretsAPI.list(workflowId || undefined));
      setLoading(false);
    })();
  }, [workflowId]);

  if (loading) return <div className="p-6">Loading…</div>;

  return (
    <div className="p-6 space-y-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Secrets</h1>

        <Link
          to="/secrets/new"
          className="px-4 py-2 bg-primary-600 text-white rounded"
        >
          Tạo Secret
        </Link>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs text-gray-600 uppercase">Tên</th>
              <th className="px-4 py-2 text-left text-xs text-gray-600 uppercase">Kiểu</th>
              <th className="px-4 py-2 text-left text-xs text-gray-600 uppercase">Workflow</th>
              <th className="px-4 py-2"></th>
            </tr>
          </thead>

          <tbody className="divide-y divide-gray-200">
            {items.map(s => (
              <tr key={s.ID}>
                <td className="px-4 py-2">{s.Name}</td>
                <td className="px-4 py-2">{s.Type}</td>
                <td className="px-4 py-2">{s.WorkflowID}</td>

                <td className="px-4 py-2 text-right space-x-3">
                  <Link
                    to={`/secrets/${s.ID}`}
                    className="text-primary-600 hover:text-primary-700"
                  >
                    Xem
                  </Link>
                </td>
              </tr>
            ))}

            {items.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center py-6 text-gray-500">
                  Không có Secrets nào được tìm thấy.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
