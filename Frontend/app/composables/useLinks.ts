import type { LinkItem } from '~/types/link'

const seed: LinkItem[] = [
  { id: '1', alias: 'kampanya-yaz', target: 'https://magaza.example.com/kampanyalar/yaz-indirimi-2026?utm_source=mail&utm_medium=bulten', clicks: 1284, locked: false, createdAt: '2026-10-02' },
  { id: '2', alias: 'sunum', target: 'https://docs.example.com/d/1x8Qe-long-document-id/edit#slide=12', clicks: 342, locked: true, createdAt: '2026-09-28' },
  { id: '3', alias: 'k7p2x', target: 'https://blog.example.com/yazilar/2026/10/linkleri-neden-kisaltmaliyiz', clicks: 96, locked: false, createdAt: '2026-09-21' },
  { id: '4', alias: 'basvuru', target: 'https://forms.example.com/r/basvuru-formu-2026-guz', clicks: 58, locked: false, createdAt: '2026-09-14' }
]

// Sahte veri deposu. API bağlanınca sadece bu dosya değişir.
export function useLinks() {
  const links = useState<LinkItem[]>('links', () => seed)

  const totalClicks = computed(() => links.value.reduce((sum, l) => sum + l.clicks, 0))

  function addLink(input: { alias: string; target: string; locked: boolean }) {
    links.value = [
      {
        id: String(Date.now()),
        alias: input.alias || Math.random().toString(36).slice(2, 7),
        target: input.target,
        clicks: 0,
        locked: input.locked,
        createdAt: new Date().toISOString().slice(0, 10)
      },
      ...links.value
    ]
  }

  function removeLink(id: string) {
    links.value = links.value.filter((l) => l.id !== id)
  }

  return { links, totalClicks, addLink, removeLink }
}
