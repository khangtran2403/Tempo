import { useEffect, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import secretsAPI from "../../api/secret";
import { Secret } from "../../types/secrets";

export default function SecretDetailPage() {
  const { id } = useParams();
  const nav = useNavigate();

  const [item, setItem] = useState<Secret | null>(null);

  useEffect(() => {
    secretsAPI.get(id!).then(setItem);
  }, [id]);

  const remove = async () => {
    // eslint-disable-next-line no-restricted-globals
    if (!confirm("Xóa secret?")) return;
    await secretsAPI.remove(id!);
    nav("/secrets");
  };

  if (!item) return <div className="p-6">Loading…</div>;

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">{item.Name}</h1>

        <div className="space-x-3">
          <Link
            to={`/secrets/${id}/edit`}
            className="px-4 py-2 bg-blue-600 text-white rounded"
          >
            Sửa
          </Link>

          <button
            onClick={remove}
            className="px-4 py-2 bg-red-600 text-white rounded"
          >
            Xóa
          </button>
        </div>
      </div>

      <div className="bg-white shadow rounded p-4">
        <pre className="text-sm bg-gray-50 p-3 rounded">
{JSON.stringify(item, null, 2)}
        </pre>
      </div>
    </div>
  );
}
