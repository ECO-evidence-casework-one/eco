import { useState } from 'react'

type MatterSummary = {
  id: string
  title: string
  status: string
  evidence_count: number
}

type WorkspaceSummary = {
  root: string
  build_id: string
  revision: number
  matter_count: number
  evidence_count: number
  preservation_count: number
  matters: MatterSummary[]
}

type BridgeAPI = {
  OpenWorkspace(root: string): Promise<void>
  CreateWorkspace(root: string): Promise<void>
  CloseWorkspace(): Promise<void>
  SnapshotSummary(): Promise<WorkspaceSummary>
  CreateMatter(title: string): Promise<void>
}

declare global {
  interface Window {
    go?: { main?: { Bridge?: BridgeAPI } }
  }
}

const api = (): BridgeAPI => {
  const bridge = window.go?.main?.Bridge
  if (!bridge) throw new Error('Wails bridge is not available. Run this spike inside Wails.')
  return bridge
}

export default function App() {
  const [root, setRoot] = useState('')
  const [matter, setMatter] = useState('')
  const [summary, setSummary] = useState<WorkspaceSummary | null>(null)
  const [status, setStatus] = useState('No workspace open')

  const run = async (operation: () => Promise<void>, success: string) => {
    try {
      await operation()
      setStatus(success)
      try { setSummary(await api().SnapshotSummary()) } catch { setSummary(null) }
    } catch (error) {
      setStatus(error instanceof Error ? error.message : String(error))
    }
  }

  return (
    <main>
      <header>
        <p className="eyebrow">NON-PRODUCTION ARCHITECTURE SPIKE</p>
        <h1>ECO · Wails v2.14 service-boundary proof</h1>
        <p>Only workspace lifecycle and matter creation are exposed. No evidence engine logic lives in this UI.</p>
      </header>

      <section aria-labelledby="workspace-heading">
        <h2 id="workspace-heading">Workspace lifecycle</h2>
        <label htmlFor="root">Workspace path</label>
        <input id="root" value={root} onChange={e => setRoot(e.target.value)} placeholder="C:\\ECO\\Workspaces\\Example" />
        <div className="actions">
          <button onClick={() => run(() => api().OpenWorkspace(root), 'Workspace opened')}>Open existing</button>
          <button onClick={() => run(() => api().CreateWorkspace(root), 'Workspace created')}>Create new</button>
          <button onClick={() => run(() => api().CloseWorkspace(), 'Workspace closed')}>Close</button>
          <button onClick={() => run(async () => { setSummary(await api().SnapshotSummary()) }, 'Summary refreshed')}>Refresh summary</button>
        </div>
      </section>

      <section aria-labelledby="matter-heading">
        <h2 id="matter-heading">Matter proof</h2>
        <label htmlFor="matter">Matter title</label>
        <input id="matter" value={matter} onChange={e => setMatter(e.target.value)} placeholder="Example matter" />
        <button onClick={() => run(() => api().CreateMatter(matter), 'Matter created')}>Create matter</button>
      </section>

      <section aria-live="polite" className="status" aria-label="Operation status">{status}</section>

      <section aria-labelledby="summary-heading">
        <h2 id="summary-heading">Safe workspace summary</h2>
        {summary ? (
          <div>
            <dl>
              <dt>Root</dt><dd>{summary.root}</dd>
              <dt>Build</dt><dd>{summary.build_id}</dd>
              <dt>Revision</dt><dd>{summary.revision}</dd>
              <dt>Matters</dt><dd>{summary.matter_count}</dd>
              <dt>Evidence</dt><dd>{summary.evidence_count}</dd>
            </dl>
            <ul>{summary.matters.map(m => <li key={m.id}>{m.title} · {m.status} · {m.evidence_count} evidence</li>)}</ul>
          </div>
        ) : <p>No summary available.</p>}
      </section>
    </main>
  )
}
