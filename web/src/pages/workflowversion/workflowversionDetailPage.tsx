import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import workflowVersionAPI, { WorkflowVersionDetail } from "../../api/workflowversion";

export default function WorkflowVersionDetailPage() {
  const { id, version } = useParams(); // workflowId + version
  const navigate = useNavigate();

  const [item, setItem] = useState<WorkflowVersionDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const data = await workflowVersionAPI.get(id!, version!);
        setItem(data);
      } catch (e: any) {
        setError(e?.message || "Thất bại khi tải phiên bản workflow");
      } finally {
        setLoading(false);
      }
    })();
  }, [id, version]);

  const activate = async () => {
    await workflowVersionAPI.activate(id!, version!);
    alert("Đã kích hoạt!");
    navigate(0); // refresh
  };

  if (loading) return <div className="p-6">Loading…</div>;
  if (error) return <div className="p-6 text-red-600">{error}</div>;
  if (!item) return <div className="p-6">Không tìm thấy</div>;

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Phiên bản {item.Version}</h1>

        {!item.IsActive && (
          <button
            onClick={activate}
            className="px-4 py-2 rounded bg-primary-600 text-white hover:bg-primary-700"
          >
            Kích hoạt phiên bản này
          </button>
        )}
      </div>

      <div className="bg-white rounded shadow p-4">
        <h2 className="text-lg font-medium mb-2">Metadata</h2>
        <pre className="bg-gray-50 p-3 rounded text-sm">
{JSON.stringify(
{
  ID: item.ID,
  WorkflowID: item.WorkflowID,
  Version: item.Version,
  CreatedAt: item.CreatedAt,
  CreatedBy: item.CreatedBy,
  ChangeSummary: item.ChangeSummary,
  IsActive: item.IsActive
},
null,
2
)}
        </pre>
      </div>

      <div className="bg-white rounded shadow p-4">
        <h2 className="text-lg font-medium mb-2">Cấu hình</h2>
        <pre className="bg-gray-50 p-3 rounded text-sm">
          {JSON.stringify(item.Definition, null, 2)}
        </pre>
      </div>
    </div>
  );
}
