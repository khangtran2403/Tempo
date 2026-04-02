import { useState } from "react";
import { useNavigate } from "react-router-dom";
import toast from "react-hot-toast";
import secretsAPI from "../../api/secret";

export default function CreateSecretPage() {
  const nav = useNavigate();

  const [name, setName] = useState("");
  const [type, setType] = useState("generic");
  const [value, setValue] = useState('{\n  "key": "value"\n}');

  const submit = async () => {
    if (!name.trim()) {
      toast.error("Name is required.");
      return;
    }

    let parsedValue;
    try {
      parsedValue = JSON.parse(value);
    } catch (e) {
      toast.error("Value is not valid JSON.");
      return;
    }

    try {
      await secretsAPI.create({
        name: name,
        type: type,
        value: parsedValue,
      } as any); // Use 'as any' to bypass strict TypeScript type checking for this one-off shape
      toast.success("Secret created successfully!");
      nav("/secrets");
    } catch (err: any) {
      toast.error(`Failed to create secret: ${err.message}`);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">Tạo Secret</h1>

      <div className="bg-white shadow rounded p-4 space-y-4">
        <div>
          <label className="block mb-1 font-medium">Name</label>
          <input
            className="border p-2 rounded w-full"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="e.g., My API Key"
          />
        </div>
        <div>
          <label className="block mb-1 font-medium">Type</label>
          <input
            className="border p-2 rounded w-full"
            value={type}
            onChange={e => setType(e.target.value)}
            placeholder="e.g., generic, api_key"
          />
        </div>
        <div>
          <label className="block mb-1 font-medium">Value (JSON)</label>
          <textarea
            className="border p-2 rounded w-full font-mono text-sm"
            rows={5}
            value={value}
            onChange={e => setValue(e.target.value)}
          />
        </div>
      </div>

      <button
        onClick={submit}
        className="px-4 py-2 bg-primary-600 text-white rounded"
      >
        Tạo
      </button>
    </div>
  );
}
