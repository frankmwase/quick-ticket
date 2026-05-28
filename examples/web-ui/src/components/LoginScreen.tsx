import { useState, type FormEvent } from 'react';
import type { Credentials } from '../api/client';
import { api } from '../api/client';

interface LoginScreenProps {
  onLogin: (creds: Credentials) => void;
}

export function LoginScreen({ onLogin }: LoginScreenProps) {
  const [baseUrl, setBaseUrl] = useState('http://localhost:8080');
  const [tenantId, setTenantId] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const creds: Credentials = { baseUrl, tenantId, apiKey };

    try {
      await api.health(creds);
      onLogin(creds);
    } catch (err) {
      setError(
        `Connection failed: ${err instanceof Error ? err.message : 'Unknown error'}`
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center h-full">
      <div className="bg-black/80 border border-terminal-green/30 backdrop-blur-sm p-8 w-full max-w-md">
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="text-terminal-green text-xs tracking-[0.5em] uppercase mb-2 opacity-60">
            ▸ CLASSIFIED ◂
          </div>
          <h1 className="text-terminal-green text-2xl font-bold tracking-wider">
            QUICK-TICKET
          </h1>
          <div className="text-terminal-green/40 text-xs mt-1 tracking-widest">
            BACKEND-AS-A-SERVICE ∕∕ TERMINAL ACCESS
          </div>
          <div className="border-b border-terminal-green/20 mt-4" />
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Base URL */}
          <div>
            <label className="block text-terminal-green/60 text-xs tracking-widest uppercase mb-1">
              ▹ ENDPOINT URL
            </label>
            <input
              id="login-base-url"
              type="text"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              className="w-full bg-transparent border border-terminal-green/30 text-terminal-green px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-green transition-colors"
              placeholder="http://localhost:8080"
            />
          </div>

          {/* Tenant ID */}
          <div>
            <label className="block text-terminal-green/60 text-xs tracking-widest uppercase mb-1">
              ▹ TENANT IDENTIFIER
            </label>
            <input
              id="login-tenant-id"
              type="text"
              value={tenantId}
              onChange={(e) => setTenantId(e.target.value)}
              className="w-full bg-transparent border border-terminal-green/30 text-terminal-green px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-green transition-colors"
              placeholder="your-tenant-id"
            />
          </div>

          {/* API Key */}
          <div>
            <label className="block text-terminal-green/60 text-xs tracking-widest uppercase mb-1">
              ▹ API KEY
            </label>
            <input
              id="login-api-key"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              className="w-full bg-transparent border border-terminal-green/30 text-terminal-green px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-green transition-colors"
              placeholder="••••••••"
            />
          </div>

          {/* Error */}
          {error && (
            <div className="text-terminal-red text-xs border border-terminal-red/30 bg-terminal-red/5 px-3 py-2">
              ⚠ {error}
            </div>
          )}

          {/* Submit */}
          <button
            id="login-submit"
            type="submit"
            disabled={loading}
            className="w-full border border-terminal-green/50 text-terminal-green py-2 text-sm tracking-widest uppercase
                       hover:bg-terminal-green/10 hover:border-terminal-green transition-all duration-300
                       disabled:opacity-30 disabled:cursor-not-allowed"
          >
            {loading ? '◉ AUTHENTICATING...' : '▸ ESTABLISH CONNECTION'}
          </button>
        </form>

        <div className="mt-6 text-center text-terminal-green/20 text-[10px] tracking-widest">
          SECURE TRANSMISSION ∕∕ ENCRYPTED CHANNEL
        </div>
      </div>
    </div>
  );
}
