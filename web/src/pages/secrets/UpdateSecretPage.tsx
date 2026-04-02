import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import secretsAPI from "../../api/secret";
import { Secret } from "../../types/secrets";

export default function EditSecretPage() {
  const { id } = useParams();
  const nav = useNavigate();

  const [item, setItem] = useState<Secret | null>(null);
  const [form, setForm] = useState<any>({});

  useEffect(() => {
    secretsAPI.get(id!).then(s => {
      setItem(s);
      setForm({
        Name: s.Name,
        Type: s.Type,
        Key: s.Key,
        Value: s.Value,
        Description: s.Description ?? ""
      });
    });
  }, [id]);

  if (!item) return <div className="p-6">Loading…</div>;

  const update = (k: string, v: string) => setForm({ ...form, [k]: v });

  const submit = async () => {
    await secretsAPI.update(id!, form);
    nav(`/secrets/${id}`);
  };

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">Chỉnh sửa Secret</h1>

      <div className="bg-white shadow rounded p-4 space-y-4">
        {Object.keys(form).map(k => (
          <div key={k}>
            <label className="block mb-1 font-medium">{k}</label>
            <input
              className="border p-2 rounded w-full"
              value={form[k]}
              onChange={e => update(k, e.target.value)}
            />
          </div>
        ))}
      </div>

      <button
        onClick={submit}
        className="px-4 py-2 bg-primary-600 text-white rounded"
      >
        Lưu thay đổi
      </button>
    </div>
  );
}
