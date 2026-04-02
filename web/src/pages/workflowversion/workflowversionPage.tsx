import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import workflowVersionAPI, {
  WorkflowVersionSummary
} from "../../api/workflowversion";

export default function WorkflowVersionsPage() {
  const { id } = useParams(); // workflowId
  const [items, setItems] = useState<WorkflowVersionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const data = await workflowVersionAPI.list(id!);
        setItems(data);
      } catch (e: any) {
        setError(e?.message || "Thất bại khi tải các phiên bản workflow");
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  if (loading) return <div className="p-6">Loading…</div>;
  if (error) return <div className="p-6 text-red-600">{error}</div>;

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-xl font-semibold">Phiên bản của Workflow {id}</h1>

        <Link
          to={`/workflows/${id}/versions/new`}
          className="px-4 py-2 bg-primary-600 text-white rounded hover:bg-primary-700"
        >
            Tạo phiên bản mới
        </Link>
      </div>

      <div className="bg-white rounded shadow">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs text-gray-500 uppercase">Phiên bản</th>
              <th className="px-4 py-2 text-left text-xs text-gray-500 uppercase">Tạo vào</th>
              <th className="px-4 py-2 text-left text-xs text-gray-500 uppercase">Hoạt động</th>
              <th className="px-4 py-2"></th>
            </tr>
          </thead>

          <tbody className="divide-y divide-gray-200">
            {items.map(v => (
              <tr key={v.ID}>
                <td className="px-4 py-2 font-mono">{v.Version}</td>
                <td className="px-4 py-2">{new Date(v.CreatedAt).toLocaleString()}</td>
                <td className="px-4 py-2">
                  {v.IsActive ? (
                    <span className="px-2 py-1 text-xs bg-green-100 text-green-700 rounded">Hoạt động</span>
                  ) : (
                    <span className="px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded">Không hoạt động</span>
                  )}
                </td>
                <td className="px-4 py-2 text-right">
                  <Link
                    className="text-primary-600 hover:text-primary-700"
                    to={`/workflows/${id}/versions/${v.Version}`}
                  >
                    Xem
                  </Link>
                    <Link
                       className="text-blue-600 hover:text-blue-700"
                        to={`/workflows/${id}/versions/compare?base=${v.Version}`}
                      >
                      So sánh giữa các phiên bản
                     </Link>
                </td>
              </tr>
            ))}

            {items.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center py-4 text-gray-500">
                  Không có phiên bản nào
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
