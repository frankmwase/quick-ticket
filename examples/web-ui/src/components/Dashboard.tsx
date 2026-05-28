import { useState, useEffect } from 'react';
import type { Credentials, UserProfile, Member, Ticket } from '../api/client';
import { api } from '../api/client';

interface DashboardProps {
  creds: Credentials;
  onLogout: () => void;
}

type Tab = 'tickets' | 'profile' | 'members';

export function Dashboard({ creds, onLogout }: DashboardProps) {
  const [tab, setTab] = useState<Tab>('tickets');
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(false);

  // Load profile on mount
  useEffect(() => {
    api.getProfile(creds).then(setProfile).catch(() => {});
    api.getMembers(creds).then(setMembers).catch(() => {});
  }, [creds]);

  // --- Ticket Actions ---
  const [genCount, setGenCount] = useState(1);
  const [genOwner, setGenOwner] = useState('');
  const [genManagedBy, setGenManagedBy] = useState('');
  const [verifyToken, setVerifyToken] = useState('');

  const handleGenerate = async () => {
    setLoading(true);
    setStatus('');
    try {
      const res = await api.generateTickets(creds, genCount, genOwner, genManagedBy);
      setTickets(res.tickets);
      setStatus(`✓ Generated ${res.count} ticket(s)`);
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : 'Failed'}`);
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async () => {
    setLoading(true);
    setStatus('');
    try {
      const res = await api.verifyTicket(creds, verifyToken);
      setStatus(res.valid ? `✓ VALID — ID: ${res.ticket_id}` : `✗ INVALID`);
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : 'Verification failed'}`);
    } finally {
      setLoading(false);
    }
  };

  // --- Profile Actions ---
  const [editName, setEditName] = useState('');
  const [editEmail, setEditEmail] = useState('');
  const [editBio, setEditBio] = useState('');

  useEffect(() => {
    if (profile) {
      setEditName(profile.full_name || '');
      setEditEmail(profile.email || '');
      setEditBio(profile.bio || '');
    }
  }, [profile]);

  const handleUpdateProfile = async () => {
    setLoading(true);
    try {
      const updated = await api.updateProfile(creds, {
        full_name: editName,
        email: editEmail,
        bio: editBio,
      });
      setProfile(updated);
      setStatus('✓ Profile updated');
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : 'Update failed'}`);
    } finally {
      setLoading(false);
    }
  };

  // --- Member Actions ---
  const [newMemberName, setNewMemberName] = useState('');
  const [newMemberRole, setNewMemberRole] = useState('operator');

  const handleCreateMember = async () => {
    if (!newMemberName) return;
    setLoading(true);
    try {
      const member = await api.createMember(creds, newMemberName, newMemberRole);
      setMembers([...members, member]);
      setNewMemberName('');
      setStatus(`✓ Member "${member.name}" created`);
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : 'Failed'}`);
    } finally {
      setLoading(false);
    }
  };

  const tabs: { id: Tab; label: string }[] = [
    { id: 'tickets', label: 'TICKETS' },
    { id: 'profile', label: 'PROFILE' },
    { id: 'members', label: 'MEMBERS' },
  ];

  return (
    <div className="h-full flex flex-col p-4 overflow-hidden">
      {/* Top Bar */}
      <div className="flex items-center justify-between border-b border-terminal-border pb-3 mb-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <span className="text-terminal-base text-lg font-bold tracking-wider">
            QT∕∕DASHBOARD
          </span>
          <span className="text-terminal-dim text-xs">
            TENANT: {creds.tenantId || 'N/A'}
          </span>
        </div>
        <button
          id="logout-btn"
          onClick={onLogout}
          className="text-terminal-red/60 text-xs border border-terminal-red/20 px-3 py-1 
                     hover:bg-terminal-red/10 hover:border-terminal-red/50 transition-all hover:animate-pulse uppercase tracking-widest"
        >
          ▸ DISCONNECT
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-0 border-b border-terminal-border/10 mb-4 flex-shrink-0">
        {tabs.map((t) => (
          <button
            key={t.id}
            id={`tab-${t.id}`}
            onClick={() => { setTab(t.id); setStatus(''); }}
            className={`px-5 py-2 text-xs tracking-widest uppercase transition-all border-b-2 ${
              tab === t.id
                ? 'text-terminal-base border-terminal-border bg-terminal-panel'
                : 'text-terminal-dim border-transparent hover:text-terminal-dim'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Status bar */}
      {status && (
        <div
          className={`text-xs px-3 py-2 mb-3 border flex-shrink-0 ${
            status.startsWith('✓')
              ? 'text-terminal-base border-terminal-border bg-terminal-panel'
              : 'text-terminal-red border-terminal-red/30 bg-terminal-red/5'
          }`}
        >
          {status}
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-auto min-h-0">
        {/* TICKETS TAB */}
        {tab === 'tickets' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Generate Panel */}
            <div className="border border-terminal-border p-5">
              <h2 className="text-terminal-amber text-xs tracking-widest uppercase mb-4">
                ▹ GENERATE TICKETS
              </h2>
              <div className="space-y-3">
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">COUNT</label>
                  <input
                    id="gen-count"
                    type="number"
                    min={1}
                    max={100}
                    value={genCount}
                    onChange={(e) => setGenCount(parseInt(e.target.value) || 1)}
                    className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                  />
                </div>
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">OWNER ID</label>
                  <input
                    id="gen-owner"
                    type="text"
                    value={genOwner}
                    onChange={(e) => setGenOwner(e.target.value)}
                    className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                    placeholder="owner-id"
                  />
                </div>
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">MANAGED BY (MEMBER)</label>
                  <select
                    id="gen-managed-by"
                    value={genManagedBy}
                    onChange={(e) => setGenManagedBy(e.target.value)}
                    className="w-full bg-black border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                  >
                    <option value="">— None —</option>
                    {members.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name} ({m.role})
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  id="gen-submit"
                  onClick={handleGenerate}
                  disabled={loading}
                  className="w-full border border-terminal-border text-terminal-base py-2 text-xs tracking-widest uppercase hover:border-terminal-accent hover:text-terminal-accent hover:shadow-[0_0_15px_rgba(0,240,255,0.2)] transition-all duration-300 transition-all disabled:opacity-30"
                >
                  {loading ? '◉ PROCESSING...' : '▸ GENERATE'}
                </button>
              </div>

              {/* Generated Tickets List */}
              {tickets.length > 0 && (
                <div className="mt-4 border-t border-terminal-border/10 pt-3">
                  <div className="text-terminal-dim text-[10px] tracking-widest mb-2">
                    GENERATED ({tickets.length})
                  </div>
                  <div className="max-h-40 overflow-auto space-y-1">
                    {tickets.map((t) => (
                      <div
                        key={t.ID}
                        className="text-[11px] text-terminal-base/70 font-mono border-l-2 border-terminal-accent border-terminal-border pl-2 py-1"
                      >
                        <span className="text-terminal-amber">ID:</span> {t.ID.substring(0, 12)}...
                        <span className="ml-2 text-terminal-dim">{t.Status}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Verify Panel */}
            <div className="border border-terminal-border p-5">
              <h2 className="text-terminal-amber text-xs tracking-widest uppercase mb-4">
                ▹ VERIFY TICKET
              </h2>
              <div className="space-y-3">
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">SECURE TOKEN</label>
                  <input
                    id="verify-token"
                    type="text"
                    value={verifyToken}
                    onChange={(e) => setVerifyToken(e.target.value)}
                    className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                    placeholder="paste-token-here"
                  />
                </div>
                <button
                  id="verify-submit"
                  onClick={handleVerify}
                  disabled={loading || !verifyToken}
                  className="w-full border border-terminal-amber/40 text-terminal-amber py-2 text-xs tracking-widest uppercase hover:bg-terminal-amber/10 transition-all disabled:opacity-30"
                >
                  {loading ? '◉ VERIFYING...' : '▸ VERIFY'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* PROFILE TAB */}
        {tab === 'profile' && (
          <div className="max-w-lg border border-terminal-border p-5">
            <h2 className="text-terminal-amber text-xs tracking-widest uppercase mb-4">
              ▹ USER PROFILE
            </h2>
            <div className="space-y-3">
              <div>
                <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">FULL NAME</label>
                <input
                  id="profile-name"
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                />
              </div>
              <div>
                <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">EMAIL</label>
                <input
                  id="profile-email"
                  type="email"
                  value={editEmail}
                  onChange={(e) => setEditEmail(e.target.value)}
                  className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                />
              </div>
              <div>
                <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">BIO</label>
                <textarea
                  id="profile-bio"
                  rows={3}
                  value={editBio}
                  onChange={(e) => setEditBio(e.target.value)}
                  className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300 resize-none"
                />
              </div>
              <button
                id="profile-save"
                onClick={handleUpdateProfile}
                disabled={loading}
                className="w-full border border-terminal-border text-terminal-base py-2 text-xs tracking-widest uppercase hover:border-terminal-accent hover:text-terminal-accent hover:shadow-[0_0_15px_rgba(0,240,255,0.2)] transition-all duration-300 transition-all disabled:opacity-30"
              >
                {loading ? '◉ SAVING...' : '▸ SAVE PROFILE'}
              </button>
            </div>
            {profile && (
              <div className="mt-4 pt-3 border-t border-terminal-border/10 text-terminal-dim text-[10px] tracking-widest">
                PROFILE ID: {profile.id}
              </div>
            )}
          </div>
        )}

        {/* MEMBERS TAB */}
        {tab === 'members' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Create Member */}
            <div className="border border-terminal-border p-5">
              <h2 className="text-terminal-amber text-xs tracking-widest uppercase mb-4">
                ▹ ADD MEMBER
              </h2>
              <div className="space-y-3">
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">NAME</label>
                  <input
                    id="member-name"
                    type="text"
                    value={newMemberName}
                    onChange={(e) => setNewMemberName(e.target.value)}
                    className="w-full bg-transparent border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                    placeholder="Agent Name"
                  />
                </div>
                <div>
                  <label className="block text-terminal-dim text-[10px] tracking-widest mb-1">ROLE</label>
                  <select
                    id="member-role"
                    value={newMemberRole}
                    onChange={(e) => setNewMemberRole(e.target.value)}
                    className="w-full bg-black border border-terminal-border text-terminal-base px-3 py-2 text-sm font-mono focus:outline-none focus:border-terminal-accent focus:shadow-[0_0_10px_rgba(0,240,255,0.2)] transition-all duration-300"
                  >
                    <option value="operator">Operator</option>
                    <option value="admin">Admin</option>
                    <option value="viewer">Viewer</option>
                  </select>
                </div>
                <button
                  id="member-create"
                  onClick={handleCreateMember}
                  disabled={loading || !newMemberName}
                  className="w-full border border-terminal-border text-terminal-base py-2 text-xs tracking-widest uppercase hover:border-terminal-accent hover:text-terminal-accent hover:shadow-[0_0_15px_rgba(0,240,255,0.2)] transition-all duration-300 transition-all disabled:opacity-30"
                >
                  {loading ? '◉ CREATING...' : '▸ CREATE MEMBER'}
                </button>
              </div>
            </div>

            {/* Members List */}
            <div className="border border-terminal-border p-5">
              <h2 className="text-terminal-amber text-xs tracking-widest uppercase mb-4">
                ▹ ACTIVE MEMBERS ({members.length})
              </h2>
              {members.length === 0 ? (
                <div className="text-terminal-base/20 text-xs">No members registered.</div>
              ) : (
                <div className="space-y-2 max-h-80 overflow-auto">
                  {members.map((m) => (
                    <div
                      key={m.id}
                      className="flex items-center justify-between border border-terminal-border/10 px-3 py-2"
                    >
                      <div>
                        <div className="text-terminal-base text-sm">{m.name}</div>
                        <div className="text-terminal-dim text-[10px] tracking-widest uppercase">
                          {m.role} ∕∕ {m.is_active ? 'ACTIVE' : 'INACTIVE'}
                        </div>
                      </div>
                      <div className="text-terminal-base/20 text-[10px]">
                        {m.id.substring(0, 8)}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
