import { useMemo, useState } from 'react'

const actions = [
  'Add evidence',
  'Review source details',
  'Create task',
  'Ask ECO',
  'Warranty confirmation',
  'Build the Matter timeline',
]

export default function App() {
  const [query, setQuery] = useState('')
  const visible = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return actions
    return actions.filter((x) => x.toLowerCase().includes(q))
  }, [query])

  return (
    <main style={{fontFamily:'system-ui, sans-serif', padding:24, maxWidth:1000, margin:'0 auto'}}>
      <h1>Synthetic Matter Workspace</h1>
      <label htmlFor="matter-search">Search this Matter</label>
      <input
        id="matter-search"
        aria-label="Search this Matter"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        style={{display:'block', width:'100%', boxSizing:'border-box', padding:10, margin:'8px 0 16px'}}
      />
      <section aria-label="Matter actions" style={{display:'flex', flexWrap:'wrap', gap:8, marginBottom:20}}>
        {visible.map((action) => <button key={action}>{action}</button>)}
      </section>
      <label htmlFor="transcript"><strong>AI conversation transcript</strong></label>
      <textarea
        id="transcript"
        aria-label="AI conversation transcript"
        readOnly
        value="Known: warranty confirmation appears in preserved source Email.eml. Source-backed synthetic qualification text only."
        style={{display:'block', width:'100%', minHeight:170, boxSizing:'border-box', marginTop:8, padding:10}}
      />
    </main>
  )
}
