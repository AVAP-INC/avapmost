import React, {useCallback, useEffect, useRef, useState} from 'react';

const PLUGIN_ID = 'com.avap.avapmost-search';
const PER_PAGE = 20;

// Mattermost globals exposed to plugins
declare global {
    interface Window {
        PostUtils?: {
            formatText(text: string, opts?: Record<string, unknown>): string;
            messageHtmlToComponent(html: string, opts?: Record<string, unknown>): React.ReactNode;
        };
    }
}

interface SearchResult {
    post_id: string;
    message: string;
    user_id: string;
    username: string;
    channel_id: string;
    channel_name: string;
    channel_type: string;
    team_name: string;
    team_slug: string;
    create_at: number;
    score: number;
}

type QueryType = 'keyword' | 'phrase' | 'semantic';

/** Mattermost 標準の日時表示 */
const formatDate = (ms: number): string => {
    if (!ms) return '';
    const d = new Date(ms);
    const now = new Date();
    const months = ['1月', '2月', '3月', '4月', '5月', '6月',
        '7月', '8月', '9月', '10月', '11月', '12月'];
    const h = d.getHours();
    const m = String(d.getMinutes()).padStart(2, '0');
    const ampm = h < 12 ? '午前' : '午後';
    const h12 = h % 12 || 12;
    const timeStr = `${ampm}${h12}:${m}`;
    const dateStr = d.getFullYear() === now.getFullYear()
        ? `${months[d.getMonth()]}${d.getDate()}日`
        : `${d.getFullYear()}年${months[d.getMonth()]}${d.getDate()}日`;
    return `${dateStr} ${timeStr}`;
};

/** メッセージを Mattermost マークダウンでレンダリング */
const renderBody = (text: string): React.ReactNode => {
    const pu = window.PostUtils;
    if (pu) {
        const formatted = pu.formatText(text, {
            mentionHighlight: false,
            markdown: true,
            singleline: false,
        });
        return pu.messageHtmlToComponent(formatted, {mentionHighlight: false});
    }
    return text;
};

const ResultCard: React.FC<{r: SearchResult}> = ({r}) => {
    const [hovered, setHovered] = useState(false);

    const navigateToPost = (e: React.MouseEvent) => {
        e.preventDefault();
        // Use team-scoped permalink to avoid team_not_found error.
        const path = r.team_slug
            ? `/${r.team_slug}/pl/${r.post_id}`
            : `/pl/${r.post_id}`;
        try {
            window.history.pushState({}, '', path);
            window.dispatchEvent(new PopStateEvent('popstate', {state: {}}));
        } catch {
            window.location.href = path;
        }
    };

    const channelPrefix = r.channel_type === 'P' ? '🔒 ' : '#';
    const channelLabel = r.team_name && r.channel_name
        ? `${r.team_name} › ${channelPrefix}${r.channel_name}`
        : (r.channel_name ? `${channelPrefix}${r.channel_name}` : r.channel_id);

    return (
        <div
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            style={{
                position: 'relative',
                padding: '16px 20px',
                borderBottom: '1px solid var(--center-channel-color-12, rgba(0,0,0,0.08))',
                background: hovered
                    ? 'var(--center-channel-color-04, rgba(0,0,0,0.04))'
                    : 'transparent',
                transition: 'background 0.1s',
            }}
        >
            <div style={{display: 'flex', gap: '10px', alignItems: 'flex-start'}}>
                {/* Avatar */}
                <div style={{flexShrink: 0, marginTop: '1px'}}>
                    {r.user_id ? (
                        <img
                            src={`/api/v4/users/${r.user_id}/image`}
                            alt=""
                            onError={(e) => {
                                const initials = (r.username || '?')[0]?.toUpperCase() ?? '?';
                                (e.target as HTMLImageElement).src =
                                    `data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32'><rect width='32' height='32' rx='16' fill='%230d5cad'/><text x='16' y='22' text-anchor='middle' fill='white' font-size='14' font-family='sans-serif'>${encodeURIComponent(initials)}</text></svg>`;
                            }}
                            style={{width: '32px', height: '32px', borderRadius: '50%', objectFit: 'cover'}}
                        />
                    ) : (
                        <div style={{
                            width: '32px', height: '32px', borderRadius: '50%',
                            background: '#0d5cad', display: 'flex', alignItems: 'center',
                            justifyContent: 'center', fontSize: '13px', fontWeight: 600, color: '#fff',
                        }}>
                            {(r.username || '?')[0]?.toUpperCase() ?? '?'}
                        </div>
                    )}
                </div>

                {/* Content */}
                <div style={{flex: 1, minWidth: 0}}>
                    {/* Header: username + timestamp */}
                    <div style={{display: 'flex', alignItems: 'baseline', gap: '8px', marginBottom: '1px', flexWrap: 'wrap'}}>
                        <span style={{fontWeight: 600, fontSize: '15px', color: 'var(--center-channel-color, #000)', lineHeight: 1.4}}>
                            @{r.username || r.user_id || '不明'}
                        </span>
                        <time style={{fontSize: '12px', color: 'var(--center-channel-color-56, #888)', lineHeight: 1.4}}>
                            {formatDate(r.create_at)}
                        </time>
                    </div>

                    {/* Channel row */}
                    <div style={{fontSize: '12px', color: 'var(--center-channel-color-56, #888)', marginBottom: '8px'}}>
                        {channelLabel}
                    </div>

                    {/* Message body */}
                    <div
                        className="post-message__text"
                        style={{fontSize: '14px', lineHeight: '1.6', color: 'var(--center-channel-color, #000)', wordBreak: 'break-word'}}
                    >
                        {renderBody(r.message)}
                    </div>
                </div>
            </div>

            {/* Jump button on hover */}
            {hovered && (
                <button
                    onClick={navigateToPost}
                    style={{
                        position: 'absolute', top: '12px', right: '16px',
                        padding: '5px 12px', borderRadius: '4px',
                        border: '1px solid var(--button-bg, #166de0)',
                        background: 'var(--center-channel-bg, #fff)',
                        color: 'var(--button-bg, #166de0)',
                        fontSize: '12px', fontWeight: 600, cursor: 'pointer', lineHeight: 1.4,
                    }}
                >
                    移動
                </button>
            )}
        </div>
    );
};

// ── Query type selector ────────────────────────────────────────────────────────

const QUERY_TYPES: {id: QueryType; label: string; title: string}[] = [
    {id: 'keyword',  label: 'キーワード',   title: 'トークナイズ検索（kuromoji）'},
    {id: 'phrase',   label: 'フレーズ',     title: '完全フレーズ一致検索'},
    {id: 'semantic', label: 'セマンティック', title: 'ベクトル類似度検索（Embedding API 要設定）'},
];

const QueryTypeSelector: React.FC<{type: QueryType; onChange(t: QueryType): void; semanticEnabled: boolean}> = ({type, onChange, semanticEnabled}) => (
    <div style={{display: 'flex', gap: '4px', alignItems: 'center'}}>
        {QUERY_TYPES.filter(({id}) => id !== 'semantic' || semanticEnabled).map(({id, label, title}) => (
            <button
                key={id}
                title={title}
                onClick={() => onChange(id)}
                style={{
                    padding: '3px 8px',
                    border: `1px solid ${type === id ? 'var(--button-bg, #166de0)' : 'var(--center-channel-color-24, #ccc)'}`,
                    borderRadius: '3px',
                    background: type === id ? 'var(--button-bg, #166de0)' : 'transparent',
                    color: type === id ? '#fff' : 'var(--center-channel-color-56, #888)',
                    fontSize: '11px',
                    fontWeight: type === id ? 600 : 400,
                    cursor: 'pointer',
                    lineHeight: 1.4,
                    transition: 'all 0.1s',
                }}
            >
                {label}
            </button>
        ))}
    </div>
);

// ── Main panel ────────────────────────────────────────────────────────────────

const SearchPanel: React.FC = () => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<SearchResult[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(0);
    const [hasMore, setHasMore] = useState(false);
    const [loading, setLoading] = useState(false);
    const [loadingMore, setLoadingMore] = useState(false);
    const [searched, setSearched] = useState(false);
    const [queryType, setQueryType] = useState<QueryType>('keyword');
    const [semanticEnabled, setSemanticEnabled] = useState(false);
    const sentinelRef = useRef<HTMLDivElement>(null);
    const scrollContainerRef = useRef<HTMLDivElement>(null);
    const queryRef = useRef('');
    const queryTypeRef = useRef<QueryType>('keyword');

    const fetchPage = useCallback(async (q: string, pageNum: number, append: boolean) => {
        try {
            const res = await fetch(`/plugins/${PLUGIN_ID}/api/v1/search`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest'},
                credentials: 'include',
                body: JSON.stringify({
                    query: q,
                    page: pageNum,
                    per_page: PER_PAGE,
                    query_type: queryTypeRef.current,
                }),
            });
            const data = await res.json() as {results?: SearchResult[]; total?: number};
            const newItems = data.results ?? [];
            setResults((prev) => append ? [...prev, ...newItems] : newItems);
            setTotal(data.total ?? 0);
            setHasMore(newItems.length === PER_PAGE);
            setPage(pageNum);
        } catch {
            if (!append) setResults([]);
            setHasMore(false);
        }
    }, []);

    const handleSearch = async () => {
        if (!query.trim()) return;
        queryRef.current = query;
        queryTypeRef.current = queryType;
        setSearched(true);
        setLoading(true);
        setResults([]);
        setHasMore(false);
        await fetchPage(query, 0, false);
        setLoading(false);
    };

    const handleQueryTypeChange = async (t: QueryType) => {
        setQueryType(t);
        if (!searched || !queryRef.current) return;
        queryTypeRef.current = t;
        setLoading(true);
        setResults([]);
        setHasMore(false);
        await fetchPage(queryRef.current, 0, false);
        setLoading(false);
    };

    const loadMore = useCallback(async () => {
        if (loadingMore || !hasMore) return;
        setLoadingMore(true);
        await fetchPage(queryRef.current, page + 1, true);
        setLoadingMore(false);
    }, [loadingMore, hasMore, page, fetchPage]);

    useEffect(() => {
        fetch(`/plugins/${PLUGIN_ID}/api/v1/config`, {credentials: 'include'})
            .then((r) => r.json() as Promise<{semantic_enabled?: boolean}>)
            .then((d) => setSemanticEnabled(d.semantic_enabled ?? false))
            .catch(() => { /* keep false */ });
    }, []);

    useEffect(() => {
        const sentinel = sentinelRef.current;
        if (!sentinel || !hasMore) return;
        const observer = new IntersectionObserver(
            ([entry]) => { if (entry.isIntersecting) void loadMore(); },
            {root: scrollContainerRef.current, rootMargin: '200px'},
        );
        observer.observe(sentinel);
        return () => observer.disconnect();
    }, [hasMore, loadMore]);

    return (
        <div style={{height: '100%', display: 'flex', flexDirection: 'column', boxSizing: 'border-box'}}>
            {/* Search input + query type selector */}
            <div style={{padding: '12px 16px', borderBottom: '1px solid var(--center-channel-color-12, rgba(0,0,0,0.08))'}}>
                <div style={{
                    display: 'flex', alignItems: 'center',
                    background: 'var(--center-channel-color-08, rgba(0,0,0,0.06))',
                    borderRadius: '4px', padding: '0 12px', gap: '8px',
                }}>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                        style={{flexShrink: 0, color: 'var(--center-channel-color-56, #888)'}}>
                        <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2"/>
                        <line x1="16.5" y1="16.5" x2="22" y2="22" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    <input
                        type="text"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') void handleSearch(); }}
                        placeholder="メッセージを検索..."
                        style={{
                            flex: 1, padding: '8px 0', border: 'none',
                            background: 'transparent', color: 'var(--center-channel-color, #000)',
                            fontSize: '14px', outline: 'none',
                        }}
                    />
                    {loading && (
                        <span style={{fontSize: '12px', color: 'var(--center-channel-color-40, #aaa)'}}>検索中…</span>
                    )}
                </div>
                <div style={{marginTop: '8px', display: 'flex', justifyContent: 'flex-end'}}>
                    <QueryTypeSelector type={queryType} onChange={handleQueryTypeChange} semanticEnabled={semanticEnabled} />
                </div>
            </div>

            {/* Results */}
            <div ref={scrollContainerRef} style={{flex: 1, overflowY: 'auto'}}>
                {!searched && (
                    <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center', marginTop: '48px', gap: '12px', padding: '0 24px'}}>
                        <svg width="64" height="64" viewBox="0 0 24 24" fill="none"
                            style={{color: 'var(--center-channel-color-24, #ccc)'}}>
                            <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="1.5"/>
                            <line x1="16.5" y1="16.5" x2="22" y2="22" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                        </svg>
                        <p style={{color: 'var(--center-channel-color-56, #888)', textAlign: 'center', fontSize: '14px', lineHeight: '1.8', margin: 0}}>
                            Elasticsearch で日本語全文検索
                        </p>
                    </div>
                )}

                {searched && !loading && results.length === 0 && (
                    <div style={{padding: '48px 24px', textAlign: 'center'}}>
                        <p style={{fontSize: '14px', color: 'var(--center-channel-color-56, #888)', margin: 0}}>
                            検索結果がありません
                        </p>
                    </div>
                )}

                {results.length > 0 && (
                    <div style={{padding: '6px 16px 4px', fontSize: '12px', color: 'var(--center-channel-color-56, #888)'}}>
                        {total} 件中 {results.length} 件表示
                    </div>
                )}

                {results.map((r) => <ResultCard key={r.post_id} r={r} />)}

                {hasMore && (
                    <div ref={sentinelRef} style={{textAlign: 'center', padding: '16px', color: 'var(--center-channel-color-40, #aaa)', fontSize: '12px'}}>
                        {loadingMore ? '読み込み中…' : ''}
                    </div>
                )}
            </div>
        </div>
    );
};

export default SearchPanel;
