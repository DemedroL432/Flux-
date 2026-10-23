import { useState, useRef, useEffect } from "react";
import { useApp } from "../context/AppContext";
import { OpenLogsFolder } from "../../wailsjs/go/main/App";
export function LogsPage() {
 const a=useApp();
 const [search,setSearch]=useState("");
 const [source,setSource]=useState("core");
 const [level,setLevel]=useState("all");
 const [follow,setFollow]=useState(true);
 const box=useRef<HTMLPreElement>(null);
 const raw=source === "core" ? a.logs.coreLogTail : a.logs.dpiLogTail;
 const lines=raw.split("\n").filter(Boolean);
 const filtered=lines.filter(l=>l.toLowerCase().includes(search.toLowerCase()) && (level==="all" || (level==="error" ? /\b(error|fatal|fail|panic)\b/i.test(l) : new RegExp("\\b"+level+"\\b","i").test(l))));
 useEffect(()=>{ if(follow && box.current) box.current.scrollTop=box.current.scrollHeight; },[raw,follow,source,a.active]);
 if(a.active!=="logs") return null;
 async function copy(){try{await navigator.clipboard.writeText(filtered.slice(-100).join("\n"));a.notify("Строки скопированы");}catch(e){a.notify("Ошибка копирования: "+String(e),"error");}}
 return <section className="concept-page"><header className="concept-heading"><div><h1>Журнал</h1><p>События sing-box и DPI · последние строки файлов</p></div></header><div className="concept-actions"><select aria-label="Источник журнала" value={source} onChange={e=>setSource(e.target.value)}><option value="core">sing-box</option><option value="dpi">Zapret / DPI</option></select><select aria-label="Уровень событий" value={level} onChange={e=>setLevel(e.target.value)}><option value="all">Все уровни</option><option value="info">INFO</option><option value="warn">WARN</option><option value="error">ERROR</option></select><input aria-label="Поиск по журналу" placeholder="Поиск по тексту" value={search} onChange={e=>setSearch(e.target.value)}/></div><article className="concept-card"><div className="concept-heading"><span>Показано {filtered.length} из {lines.length}</span><label><input type="checkbox" checked={follow} onChange={e=>setFollow(e.target.checked)}/> Автопрокрутка</label></div><pre ref={box} className="concept-log" tabIndex={0}>{filtered.join("\n") || "Нет записей по выбранному фильтру"}</pre></article><div className="concept-actions"><button className="concept-button" disabled={a.busy} onClick={()=>void a.refreshLogs()}>Обновить</button><button className="concept-button" disabled={!filtered.length} onClick={()=>void copy()}>Копировать 100 строк</button><button className="concept-button" onClick={()=>void OpenLogsFolder().catch(e=>a.notify(String(e),"error"))}>Папка логов</button><button className="concept-button" disabled={a.busy} onClick={()=>void a.clearLogFiles()}>Очистить</button></div></section>;
}
