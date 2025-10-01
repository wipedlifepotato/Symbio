import React, { useEffect, useMemo, useState } from "react";

// Tailwind-first single-file app. Mobile-first with bottom nav + FAB; desktop workbench.
export default function App(){
  // -------- State
  const [page,setPage] = useState("dashboard");
  const [stack,setStack] = useState(["dashboard"]);
  const [lang,setLang] = useState("uk");
  const [toast,setToast] = useState("");
  const [drawer,setDrawer] = useState({open:false,title:"",content:null});
  const isMobile = typeof window!=="undefined" && window.matchMedia && window.matchMedia("(max-width: 767px)").matches;

  // -------- I18N (compact)
  const t = useMemo(()=>({
    uk:{brand:"FX", tabs:{dash:"Dashboard", jobs:"Задачі", disp:"Диспути", adm:"Адмінка"},
        heroTitle:"Ласкаво просимо", heroText:"Швидкий доступ до ескроу, задач і диспутів.", ctaCreate:"＋ Створити", ctaBrowse:"Переглянути задачі",
        escrowTitle:"Ескроу-баланс", held:"Утримується", releasable:"Доступно", pending:"Очікує", payout:"Запит відправлено",
        finances:"Фінанси", history:"Історія виплат", disputes:"Диспути", system:"Система фрілансу",
        allJobs:"Всі задачі", createJob:"Створити задачу", admin:"Адмінка", back:"Назад",
        filters:"Фільтри", escrowOnly:"Тільки ескроу", urgentOnly:"Термінові", details:"Деталі", chooseJob:"Обери задачу зі списку.",
        criteria:"Критерії приймання", apply:"Відгукнутись", ask:"Поставити питання", report:"Репорт", ago:"тому",
        dfOpen:"Відкриті", dfRev:"На розгляді", dfRes:"Закриті", noCases:"Немає кейсів.",
        kUsers:"Користувачі", kActive:"Активні задачі", kDisp:"Диспути", kRev:"Дохід (30д)",
        mUsers:"Модерація користувачів", block:"Блокувати", unblock:"Розблокувати", verify:"Верифікувати",
        mJobs:"Модерація задач", approve:"Схвалити", reject:"Відхилити", republish:"Переопублікувати",
        addEvidence:"Додати доказ", openArb:"Відкрити арбітраж", closeCase:"Закрити кейс"},
    ru:{brand:"FX", tabs:{dash:"Панель", jobs:"Задачи", disp:"Диспуты", adm:"Админка"},
        heroTitle:"Добро пожаловать", heroText:"Быстрый доступ к эскроу, задачам и диспутам.", ctaCreate:"＋ Создать", ctaBrowse:"Смотреть задачи",
        escrowTitle:"Баланс эскроу", held:"Удерживается", releasable:"Доступно", pending:"Ожидает", payout:"Запрос отправлен",
        finances:"Финансы", history:"История выплат", disputes:"Диспуты", system:"Система фриланса",
        allJobs:"Все задачи", createJob:"Создать задачу", admin:"Админка", back:"Назад",
        filters:"Фильтры", escrowOnly:"Только эскроу", urgentOnly:"Срочные", details:"Детали", chooseJob:"Выберите задачу из списка.",
        criteria:"Критерии приёмки", apply:"Откликнуться", ask:"Задать вопрос", report:"Репорт", ago:"назад",
        dfOpen:"Открытые", dfRev:"На рассмотрении", dfRes:"Закрытые", noCases:"Нет кейсов.",
        kUsers:"Пользователи", kActive:"Активные задачи", kDisp:"Диспуты", kRev:"Доход (30д)",
        mUsers:"Модерация пользователей", block:"Заблокировать", unblock:"Разблокировать", verify:"Верифицировать",
        mJobs:"Модерация задач", approve:"Одобрить", reject:"Отклонить", republish:"Переопубликовать",
        addEvidence:"Добавить доказательство", openArb:"Открыть арбитраж", closeCase:"Закрыть кейс"},
    en:{brand:"FX", tabs:{dash:"Dashboard", jobs:"Jobs", disp:"Disputes", adm:"Admin"},
        heroTitle:"Welcome", heroText:"Quick access to escrow, jobs and disputes.", ctaCreate:"＋ Create", ctaBrowse:"Browse jobs",
        escrowTitle:"Escrow balance", held:"Held", releasable:"Releasable", pending:"Pending", payout:"Payout requested",
        finances:"Finance", history:"Payout history", disputes:"Disputes", system:"Freelance system",
        allJobs:"All jobs", createJob:"Create job", admin:"Admin", back:"Back",
        filters:"Filters", escrowOnly:"Escrow only", urgentOnly:"Urgent", details:"Job details", chooseJob:"Select a job.",
        criteria:"Acceptance criteria", apply:"Apply", ask:"Ask question", report:"Report", ago:"ago",
        dfOpen:"Open", dfRev:"In review", dfRes:"Resolved", noCases:"No cases.",
        kUsers:"Users", kActive:"Active jobs", kDisp:"Disputes", kRev:"Revenue (30d)",
        mUsers:"User moderation", block:"Block", unblock:"Unblock", verify:"Verify",
        mJobs:"Job moderation", approve:"Approve", reject:"Reject", republish:"Republish",
        addEvidence:"Add evidence", openArb:"Open arbitration", closeCase:"Close case"}
  }),[]);
  const L = (k)=>t[lang][k];

  // -------- Data
  const tags = ["react","node","python","design","telegram"];
  const [activeTags,setActiveTags] = useState([]);
  const [fEsc,setFEsc] = useState(false); const [fUrg,setFUrg] = useState(false);
  const jobs = [
    {id:101, title:"Лендінг для кавʼярні", budget:700, cur:"USD", ddl:7, tags:["design","react"], posted:"3h", desc:"Лендінг з оплатою, адаптив, 5 секцій.", crit:["Pixel-perfect","Оплата Stripe"]},
    {id:102, title:"Телеграм-бот для заявок", budget:500, cur:"USD", ddl:5, tags:["python","telegram"], posted:"1h", desc:"Бот заявки, адмінка, експорт.", crit:["Антиспам","Ролі"]},
    {id:103, title:"Dashboard React + Node", budget:20, cur:"USD/h", ddl:14, tags:["react","node"], posted:"12h", desc:"Адмінка: таблиці, фільтри.", crit:["JWT","Unit tests 60%"]}
  ];
  const disputes = [
    {id:"D-201", job:"#101", status:"open", updated:"2h", title:"Невідповідність ТЗ", desc:"Клієнт вважає, що макет не відповідає."},
    {id:"D-202", job:"#099", status:"review", updated:"1d", title:"Прострочений дедлайн", desc:"Виконавець просить +3 дні."},
    {id:"D-203", job:"#087", status:"resolved", updated:"4d", title:"Спір щодо виплати", desc:"Ескроу розблоковано."}
  ];
  const sLabel = (s)=> lang==="ru" ? (s==="open"?"Открыт":s==="review"?"На рассмотрении":"Закрыт") : lang==="en" ? (s==="open"?"Open":s==="review"?"In review":"Resolved") : (s==="open"?"Відкритий":s==="review"?"На розгляді":"Закритий");

  // -------- Effects
  useEffect(()=>{ const saved = localStorage.getItem("fx-lang"); if(saved) setLang(saved); },[]);
  useEffect(()=>{ localStorage.setItem("fx-lang", lang); },[lang]);

  // -------- Nav helpers
  const go = (p)=>{ setPage(p); setStack((s)=>[...s,p]); };
  const back = ()=>{ setStack((s)=>{ if(s.length>1){ const ns=[...s]; ns.pop(); setPage(ns[ns.length-1]); return ns; } setPage("dashboard"); return ["dashboard"];}); };

  // -------- UI helpers
  const Toast = ()=> toast ? (
    <div className="fixed left-1/2 -translate-x-1/2 bottom-4 translate-y-0 bg-white border border-slate-200 shadow-xl rounded-xl px-3 py-2 text-sm z-[110] transition">
      {toast}
    </div>
  ) : null;
  const pushToast = (msg)=>{ setToast(msg); setTimeout(()=>setToast(""), 1400); };
  const openDrawerWith = (title,content)=> setDrawer({open:true,title,content});
  const closeDrawer = ()=> setDrawer({open:false,title:"",content:null});

  // -------- Reusable bits
  const KPI = ({label,value})=> (
    <div className="bg-white border border-slate-200 rounded-xl p-4 shadow">
      <div className="text-slate-500 text-sm">{label}</div>
      <div className="text-2xl font-extrabold">{value}</div>
    </div>
  );
  const Tile = ({title,desc,onClick})=> (
    <article className="bg-white border border-slate-200 rounded-2xl p-5 shadow cursor-pointer" onClick={onClick}>
      <div className="font-semibold">{title}</div>
      <p className="text-slate-500 text-sm mt-1">{desc}</p>
    </article>
  );

  // -------- Pages
  const Dashboard = (
    <section className="space-y-4">
      <div className="grid md:grid-cols-2 gap-4">
        <article className="bg-white border border-slate-200 rounded-xl p-4 shadow">
          <h3 className="font-semibold mb-2">{L("escrowTitle")}</h3>
          <div className="grid grid-cols-3 gap-3">
            <div><div className="text-slate-500 text-sm">{L("held")}</div><div className="font-bold">$1,240</div></div>
            <div><div className="text-slate-500 text-sm">{L("releasable")}</div><div className="font-bold">$560</div></div>
            <div><div className="text-slate-500 text-sm">{L("pending")}</div><div className="font-bold">$180</div></div>
          </div>
          <button className="mt-3 bg-white border border-slate-200 rounded-xl px-3 py-2 shadow font-bold" onClick={()=>pushToast(L("payout"))}>{L("history")}</button>
        </article>
        <article className="bg-white border border-slate-200 rounded-xl p-4 shadow">
          <h3 className="font-semibold mb-2">{L("finances")}</h3>
          <div className="flex gap-2 flex-wrap">
            <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast(L("history"))}>{L("history")}</button>
            <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>go("disputes")}>{L("disputes")}</button>
          </div>
        </article>
      </div>
      <h3 className="font-semibold">{L("system")}</h3>
      <div className="grid md:grid-cols-4 gap-4">
        <Tile title={L("allJobs")} desc="Стрічка замовлень з фільтрами" onClick={()=>go("jobs")} />
        <Tile title={L("createJob")} desc="Два кліки і поїхали" onClick={()=>pushToast("Create modal (demo)")} />
        <Tile title={L("disputes")} desc="Арбітраж за 48 год" onClick={()=>go("disputes")} />
        <Tile title={L("admin")} desc="KPI та модерація" onClick={()=>go("admin")} />
      </div>
    </section>
  );

  const JobDetails = ({job})=> (
    <div>
      <p>{job.desc}</p>
      <h4 className="font-semibold mt-2">{L("criteria")}</h4>
      <ul className="list-disc pl-5 text-sm">
        {job.crit.map((c,i)=>(<li key={i}>{c}</li>))}
      </ul>
      <div className="flex gap-2 flex-wrap mt-3">
        <button className="bg-sky-500 text-white rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast(L("apply"))}>{L("apply")}</button>
        <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast(L("ask"))}>{L("ask")}</button>
        <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast(L("report"))}>{L("report")}</button>
      </div>
    </div>
  );

  const Workbench = () => {
    const filtered = jobs.filter(j => (activeTags.length? activeTags.some(t=>j.tags.includes(t)) : true) && (!fEsc || true) && (!fUrg || false));
    const openJob = (j)=>{
      if(isMobile) openDrawerWith(L("details"), <JobDetails job={j}/>);
      else setSide(<JobDetails job={j}/>);
    };
    const [side,setSide] = useState(null);
    useEffect(()=>{ setSide(null); },[activeTags,fEsc,fUrg]);
    return (
      <section>
        <div className="flex items-center gap-2 mb-3"><button className="border border-slate-200 rounded-lg px-3 py-1" onClick={back}>← {L("back")}</button><h2 className="font-bold text-lg">{L("allJobs")}</h2></div>
        <div className="grid gap-3 lg:grid-cols-[260px_1fr_360px]">
          <aside className="bg-white border border-slate-200 rounded-xl p-4 shadow">
            <h3 className="font-semibold mb-2">{L("filters")}</h3>
            <div className="flex gap-2 flex-wrap">{tags.map(tag=> (
              <span key={tag} className={`px-2 py-1 rounded-full text-sm font-semibold cursor-pointer ${activeTags.includes(tag)?"outline outline-sky-500": "bg-slate-100"}`} onClick={()=> setActiveTags(p=> p.includes(tag)? p.filter(x=>x!==tag): [...p,tag])}>{tag}</span>
            ))}</div>
            <label className="flex items-center gap-2 mt-3 text-sm"><input type="checkbox" checked={fEsc} onChange={e=>setFEsc(e.target.checked)}/> {L("escrowOnly")}</label>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={fUrg} onChange={e=>setFUrg(e.target.checked)}/> {L("urgentOnly")}</label>
          </aside>
          <section className="grid gap-3">
            {filtered.map(j => (
              <article key={j.id} className="bg-white border border-slate-200 rounded-xl p-3 shadow cursor-pointer" onClick={()=>openJob(j)}>
                <div className="flex justify-between"><strong>{j.title}</strong><span>{j.budget} {j.cur}</span></div>
                <div className="text-slate-500 text-xs">⏱ {j.ddl}d • 📌 {j.posted} {L("ago")}</div>
                <div className="mt-1 flex gap-1 flex-wrap">{j.tags.map(t => <span className="bg-slate-100 rounded-full px-2 py-0.5 text-[12px] font-bold" key={t}>{t}</span>)}</div>
              </article>
            ))}
          </section>
          <aside className="bg-white border border-slate-200 rounded-xl p-4 shadow hidden lg:block min-h-[160px]">
            <h3 className="font-semibold mb-2">{L("details")}</h3>
            {!side ? <p className="text-slate-500 text-sm">{L("chooseJob")}</p> : side}
          </aside>
        </div>
      </section>
    );
  };

  const Disputes = () => {
    const [o,setO]=useState(true),[r,setR]=useState(true),[s,setS]=useState(true);
    const items = disputes.filter(d => (d.status==="open"&&o)||(d.status==="review"&&r)||(d.status==="resolved"&&s));
    const openCase = (d)=>{
      const tl=[{time:"10:12",actor:"Клієнт",text:"Відкрив диспут."},{time:"10:40",actor:"Виконавець",text:"Надав оновлений макет."},{time:"11:05",actor:"Арбітраж",text:"Попросив додаткові файли."}];
      openDrawerWith(L("disputes"), (
        <div className="space-y-3">
          {tl.map((e,i)=> (
            <div key={i} className="grid grid-cols-[76px_1fr] gap-3"><time className="text-slate-500 text-xs">{e.time}</time><div><strong>{e.actor}</strong><div>{e.text}</div></div></div>
          ))}
          <div className="flex gap-2 flex-wrap">
            <button className="bg-sky-500 text-white rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast("OK")}>+ {L("addEvidence")}</button>
            <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast("OK")}>{L("openArb")}</button>
            <button className="border border-slate-200 rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast("OK")}>{L("closeCase")}</button>
          </div>
        </div>
      ));
    };
    return (
      <section>
        <div className="flex items-center gap-2 mb-3"><button className="border border-slate-200 rounded-lg px-3 py-1" onClick={back}>← {L("back")}</button><h2 className="font-bold text-lg">{L("disputes")}</h2></div>
        <div className="grid md:grid-cols-2 gap-3">
          <aside className="bg-white border border-slate-200 rounded-xl p-4 shadow space-y-1">
            <h3 className="font-semibold mb-2">{L("filters")}</h3>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={o} onChange={e=>setO(e.target.checked)}/> {L("dfOpen")}</label>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={r} onChange={e=>setR(e.target.checked)}/> {L("dfRev")}</label>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={s} onChange={e=>setS(e.target.checked)}/> {L("dfRes")}</label>
          </aside>
          <section className="bg-white border border-slate-200 rounded-xl p-4 shadow">
            {items.map(d => (
              <article key={d.id} className="border border-slate-200 rounded-lg p-3 mb-3 cursor-pointer" onClick={()=>openCase(d)}>
                <div className="flex justify-between"><strong>{d.title}</strong><span>{d.job}</span></div>
                <div className="text-slate-500 text-xs">⏱ {d.updated} {L("ago")} • {sLabel(d.status)}</div>
                <p className="mt-1 text-sm">{d.desc}</p>
              </article>
            ))}
            {!items.length && <p className="text-slate-500 text-sm">{L("noCases")}</p>}
          </section>
        </div>
      </section>
    );
  };

  const Admin = (
    <section>
      <div className="flex items-center gap-2 mb-3"><button className="border border-slate-200 rounded-lg px-3 py-1" onClick={back}>← {L("back")}</button><h2 className="font-bold text-lg">{L("admin")}</h2></div>
      <div className="grid md:grid-cols-2 gap-3">
        <KPI label={L("kUsers")} value="1,248"/>
        <KPI label={L("kActive")} value="312"/>
        <KPI label={L("kDisp")} value="7"/>
        <KPI label={L("kRev")} value="$18.4k"/>
      </div>
      <div className="grid md:grid-cols-2 gap-3 mt-3">
        <article className="bg-white border border-slate-200 rounded-xl p-4 shadow space-x-2">
          <h3 className="font-semibold mb-2">{L("mUsers")}</h3>
          <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={()=>pushToast("OK")}>{L("block")}</button>
          <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={()=>pushToast("OK")}>{L("unblock")}</button>
          <button className="bg-sky-500 text-white rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast("OK")}>{L("verify")}</button>
        </article>
        <article className="bg-white border border-slate-200 rounded-xl p-4 shadow space-x-2">
          <h3 className="font-semibold mb-2">{L("mJobs")}</h3>
          <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={()=>pushToast("OK")}>{L("approve")}</button>
          <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={()=>pushToast("OK")}>{L("reject")}</button>
          <button className="bg-sky-500 text-white rounded-lg px-3 py-1 font-semibold" onClick={()=>pushToast("OK")}>{L("republish")}</button>
        </article>
      </div>
    </section>
  );

  // -------- Layout
  return (
    <div className="min-h-dvh bg-slate-50 text-slate-900">
      {/* Topbar */}
      <header className="sticky top-0 z-40 bg-white/80 backdrop-blur border-b border-slate-200">
        <div className="max-w-[1100px] mx-auto px-4 h-14 grid grid-cols-[1fr_auto_auto] items-center gap-3">
          <div className="flex items-center gap-2 font-extrabold"><div className="w-6 h-6 rounded-lg bg-gradient-to-br from-sky-300 to-blue-600"/> {t[lang].brand}</div>
          <nav className="hidden sm:flex gap-1">
            {(["dashboard","jobs","disputes","admin"]).map(p => (
              <button key={p} onClick={()=>go(p)} className={`border border-slate-200 rounded-lg px-2.5 py-1.5 font-bold ${page===p?"bg-sky-500 text-white border-sky-600":"bg-white"}`}>{t[lang].tabs[p==="dashboard"?"dash":p==="jobs"?"jobs":p==="disputes"?"disp":"adm"]}</button>
            ))}
          </nav>
          <div className="flex items-center gap-1">
            {(["uk","ru"]).map(l => (
              <span key={l} onClick={()=>setLang(l)} className={`px-2 py-1 rounded-full text-xs font-bold cursor-pointer ${lang===l?"outline outline-sky-500":"bg-slate-100"}`}>{l.toUpperCase()}</span>
            ))}
          </div>
        </div>
      </header>

      {/* Hero (dashboard only) */}
      {page==="dashboard" && (
        <section className="max-w-[1100px] mx-auto px-4 mt-3">
          <div className="bg-gradient-to-br from-sky-400 to-blue-600 text-white rounded-2xl shadow p-5 flex justify-between gap-3">
            <div>
              <h1 className="text-xl font-extrabold">{L("heroTitle")}</h1>
              <p className="text-white/80 text-sm">{L("heroText")}</p>
            </div>
            <div className="flex gap-2 flex-wrap">
              <button className="bg-white text-slate-900 rounded-xl px-3 py-2 font-extrabold shadow" onClick={()=>pushToast("Create modal (demo)")}>{L("ctaCreate")}</button>
              <button className="border border-white/60 rounded-xl px-3 py-2 font-extrabold" onClick={()=>go("jobs")}>{L("ctaBrowse")}</button>
            </div>
          </div>
        </section>
      )}

      <main className="max-w-[1100px] mx-auto px-4 py-4 space-y-4">
        {page==="dashboard" && Dashboard}
        {page==="jobs" && <Workbench/>}
        {page==="disputes" && <Disputes/>}
        {page==="admin" && Admin}
      </main>

      {/* Bottom nav (mobile) */}
      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 shadow-[0_-10px_25px_rgba(2,8,23,0.06)] grid grid-cols-4 px-2 py-1 sm:hidden">
        <button onClick={()=>go("dashboard")} className={`py-1 font-bold ${page==="dashboard"?"text-slate-900":"text-slate-500"}`}>🏠<div className="text-[11px]">Головна</div></button>
        <button onClick={()=>go("jobs")} className={`py-1 font-bold ${page==="jobs"?"text-slate-900":"text-slate-500"}`}>🧩<div className="text-[11px]">{L("allJobs")}</div></button>
        <button onClick={()=>go("disputes")} className={`py-1 font-bold ${page==="disputes"?"text-slate-900":"text-slate-500"}`}>⚖️<div className="text-[11px]">{L("disputes")}</div></button>
        <button onClick={()=>go("admin")} className={`py-1 font-bold ${page==="admin"?"text-slate-900":"text-slate-500"}`}>🛠<div className="text-[11px]">{L("admin")}</div></button>
      </nav>
      <button onClick={()=>pushToast("Create modal (demo)")} className="fixed right-4 bottom-20 sm:hidden w-14 h-14 rounded-full bg-sky-500 text-white text-3xl shadow">＋</button>

      {/* Drawer */}
      {drawer.open && (
        <>
          <div className="fixed inset-0 bg-black/40" onClick={closeDrawer}/>
          <div className="fixed left-0 right-0 bottom-0 bg-white rounded-t-2xl shadow max-h-[85vh] grid grid-rows-[auto_1fr]">
            <div className="flex items-center justify-between gap-2 px-4 py-3 border-b border-slate-200">
              <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={closeDrawer}>←</button>
              <strong>{drawer.title}</strong>
              <button className="border border-slate-200 rounded-lg px-3 py-1" onClick={closeDrawer}>✕</button>
            </div>
            <div className="overflow-auto p-4">{drawer.content}</div>
          </div>
        </>
      )}

      <Toast/>
    </div>
  );
}
