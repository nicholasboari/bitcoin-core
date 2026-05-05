const ids = {
  refreshStatus: document.querySelector("#refresh-status"),
  eventRate: document.querySelector("#event-rate"),
  txObserved: document.querySelector("#tx-observed"),
  blocksObserved: document.querySelector("#blocks-observed"),
  eventState: document.querySelector("#event-state"),
  divergenceCard: document.querySelector("#divergence-card"),
  divergenceValue: document.querySelector("#divergence-value"),
  bestBlock: document.querySelector("#best-block"),
  lastSeenBlock: document.querySelector("#last-seen-block"),
  txList: document.querySelector("#tx-list"),
  blockList: document.querySelector("#block-list"),
};

const summaryWindowSeconds = 60;
const numberFormat = new Intl.NumberFormat("pt-BR");
const rateFormat = new Intl.NumberFormat("pt-BR", {
  maximumFractionDigits: 2,
});

async function fetchJSON(url) {
  const response = await fetch(url, { headers: { Accept: "application/json" } });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `HTTP ${response.status}`);
  }

  return response.json();
}

function setText(element, value) {
  element.textContent = value;
}

function shortHash(value) {
  if (!value) {
    return "--";
  }

  if (value.length <= 18) {
    return value;
  }

  return `${value.slice(0, 10)}...${value.slice(-8)}`;
}

function formatTime(ts) {
  if (!ts) {
    return "--";
  }

  return new Date(ts * 1000).toLocaleTimeString("pt-BR");
}

function updateSummary(summary) {
  const txObserved = summary.tx_observed || 0;
  const blocksObserved = summary.blocks_observed || 0;
  const eventRate = (txObserved + blocksObserved) / summaryWindowSeconds;

  setText(ids.eventRate, `${rateFormat.format(eventRate)}/s`);
  setText(ids.txObserved, numberFormat.format(txObserved));
  setText(ids.blocksObserved, numberFormat.format(blocksObserved));

  if (txObserved + blocksObserved === 0) {
    ids.eventState.classList.remove("active");
    setText(ids.eventState, "Nenhum evento nos últimos 60s");
    return;
  }

  ids.eventState.classList.add("active");
  setText(ids.eventState, `Último evento às ${formatTime(summary.last_event_time)}`);
}

function updateDivergence(comparison) {
  const divergent = Boolean(comparison.divergence);

  ids.divergenceCard.classList.toggle("divergent", divergent);
  ids.divergenceCard.classList.toggle("synced", !divergent);

  setText(ids.divergenceValue, divergent ? "Divergente" : "Sincronizado");
  setText(ids.bestBlock, shortHash(comparison.best_block));
  setText(ids.lastSeenBlock, shortHash(comparison.last_seen_block));
  ids.bestBlock.title = comparison.best_block || "";
  ids.lastSeenBlock.title = comparison.last_seen_block || "";
}

function renderEvents(list, events, hashKey, emptyMessage) {
  list.innerHTML = "";

  if (!events || events.length === 0) {
    const item = document.createElement("li");
    item.className = "empty-state";
    item.textContent = emptyMessage;
    list.append(item);
    return;
  }

  for (const event of events) {
    const item = document.createElement("li");
    const hash = event[hashKey] || "";
    const hashElement = document.createElement("span");
    const timeElement = document.createElement("time");

    hashElement.className = "event-hash";
    hashElement.title = hash;
    hashElement.textContent = shortHash(hash);
    timeElement.textContent = formatTime(event.ts);

    item.append(hashElement, timeElement);
    list.append(item);
  }
}

function updateLatest(latest) {
  renderEvents(ids.txList, latest.txs, "txid", "Nenhuma transação observada");
  renderEvents(ids.blockList, latest.blocks, "hash", "Nenhum bloco observado");
}

async function refreshDashboard() {
  ids.refreshStatus.classList.remove("error");
  setText(ids.refreshStatus, "Atualizando...");

  try {
    const [summary, latest, comparison] = await Promise.all([
      fetchJSON(`/api/events/summary?seconds=${summaryWindowSeconds}`),
      fetchJSON("/api/events/latest"),
      fetchJSON("/api/events/state-comparison"),
    ]);

    updateSummary(summary);
    updateLatest(latest);
    updateDivergence(comparison);
    setText(ids.refreshStatus, `Atualizado ${new Date().toLocaleTimeString("pt-BR")}`);
  } catch (error) {
    ids.refreshStatus.classList.add("error");
    ids.eventState.classList.remove("active");
    setText(ids.refreshStatus, "Erro ao atualizar");
    setText(ids.eventState, error.message.trim());
  }
}

refreshDashboard();
setInterval(refreshDashboard, 5_000);
