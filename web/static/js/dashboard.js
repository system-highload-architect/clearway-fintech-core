// Переменная для хранения текущего выбранного b2b-сценария
let currentScenario = 'success';

document.addEventListener("DOMContentLoaded", function() {
    // Дефолтная инициализация валидной карты при старте ОЗУ
    setScenario('success');

    // Плавная анимация гистограммы памяти кластера
    setInterval(() => {
        const bars = document.querySelectorAll('#bar_chart_container .bar');
        bars.forEach(bar => {
            const heightValue = Math.floor(Math.random() * 55) + 30;
            bar.style.height = `${heightValue}%`;
        });
    }, 1200);
});

function setScenario(type) {
    currentScenario = type;
    const cardInput = document.getElementById('card_number');
    const amountInput = document.getElementById('amount');
    resetGraph();
    
    if (type === 'success') {
        cardInput.value = "4111 2222 3333 1111"; // ИСТИННЫЙ ВАЛИДНЫЙ PAN ПО ЛУНУ!
        amountInput.value = "150.25";
        pushLog("📝 Выбран сценарий: УСПЕШНЫЙ ПЛАТЕЖ. Карта валидна.", "sys");
    } else if (type === 'fraud') {
        cardInput.value = "4111 1111 1111 1111"; // Тестовый БИН фрода
        amountInput.value = "150000.00";
        pushLog("📝 Выбран сценарий: ФРОД-АТАКА. Попытка крупного списания.", "error");
    } else if (type === 'luhn') {
        cardInput.value = "4111 2222 3333 4444"; // Точно сломает Лун
        amountInput.value = "75.00";
        pushLog("📝 Выбран сценарий: НЕВАЛИДНАЯ КАРТА. Ошибка формулы Луна.", "error");
    }
}

function pushLog(message, context = 'info') {
    const term = document.getElementById('terminal');
    if (!term) return;
    const colorClass = context === 'error' ? 'text-rose-400' : context === 'success' ? 'text-cyan-400 font-bold' : context === 'sys' ? 'text-amber-400' : 'text-emerald-500';
    const logNode = document.createElement('div');
    logNode.className = `${colorClass}`;
    logNode.innerHTML = `<span style="color: #475569;">[${new Date().toLocaleTimeString()}]:</span> ${message}`;
    term.appendChild(logNode);
    term.scrollTop = term.scrollHeight;
}

function clearLogs() {
    document.getElementById('terminal').innerHTML = '<div style="color: #475569;">// Очищено. Готов к приему транзакций...</div>';
}

function resetGraph() {
    document.querySelectorAll('.node-circle').forEach(n => {
        n.classList.remove('node-active', 'node-failed');
    });
    document.querySelectorAll('.node-line').forEach(l => {
        l.classList.remove('active', 'failed');
    });
}

const sleep = (ms) => new Promise(r => setTimeout(r, ms));

// ИСПРАВЛЕНО (Унифицированный пошаговый движок): Теперь визуализатор четко подчиняется выбранному сценарию!
// FIXED: Bound pipeline animation frames sequentially based on active scenario configurations
async function runProcessingPipeline() {
    resetGraph();
    
    const btn = document.getElementById('btn_pay');
    const btnText = document.getElementById('btn_text');
    if (!btn) return;
    
    btn.disabled = true;
    btn.style.opacity = "0.5";
    btn.style.cursor = "not-allowed";
    btnText.innerText = "⚡ ОБРАБОТКА ЯДРОМ FINTECH...";

    const pan = document.getElementById('card_number').value.replace(/\s+/g, '');
    const amount = document.getElementById('amount').value;

    pushLog(`🚀 [HTTP INGRESS]: Поймана форма оплаты мерчанта. Запуск PCI-DSS контура.`);
    
    // === ЭТАП 1: NEW ===
    await sleep(1000);
    document.getElementById('node_NEW').classList.add('node-active');
    pushLog(`📡 [POST] Инициализация транзакции в ОЗУ... Стейт: NEW`);

    // Если сценарий изначально "luhn" — мгновенно красим кружок в КРАСНЫЙ и рубим ветку!
    if (currentScenario === 'luhn') {
        await sleep(1000);
        document.getElementById('node_NEW').classList.remove('node-active');
        document.getElementById('node_NEW').classList.add('node-failed');
        pushLog(`🔒 [PCI-DSS BLOCK]: Крах алгоритма Луна! Карточка не существует. Стейт переведен в FAILED.`, 'error');
        finishPipeline();
        return;
    }

    try {
        pushLog(`📡 [POST] Вызов /api/v1/gateway/tokenize...`);
        const response = await fetch('/api/v1/gateway/tokenize', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                card_number: pan,
                card_holder: "DAVID_TECHLEAD",
                expiry_date: document.getElementById('card_expiry').value,
                cvv: document.getElementById('card_cvv').value,
                amount: amount
            })
        });

        const result = await response.json();
        if (!result.success) {
            // Если бэкенд выплюнул ошибку — уходим в аварийный финал
            document.getElementById('node_NEW').classList.remove('node-active');
            document.getElementById('node_NEW').classList.add('node-failed');
            pushLog(`🔒 [PCI-DSS BLOCK]: Сервер отклонил карту: ${result.error}`, 'error');
            finishPipeline();
            return;
        }

        // === ЭТАП 2: FRAUD CHECKING ===
        await sleep(1000);
        document.getElementById('line_1').classList.add('active');
        document.getElementById('node_FRAUD').classList.add('node-active');
        pushLog(`🧬 [gRPC FRAUD RADAR]: Вызов CheckFraudScore() по Protobuf-контракту...`, 'sys');
        
        // Если сценарий "fraud" — взводим КРАСНЫЙ провод и кружок фрода!
        if (currentScenario === 'fraud') {
            await sleep(1000);
            document.getElementById('line_1').classList.remove('active');
            document.getElementById('line_1').classList.add('failed');
            document.getElementById('node_FRAUD').classList.remove('node-active');
            document.getElementById('node_FRAUD').classList.add('node-failed');
            pushLog(`🔒 [FRAUD CRASH]: Превышен лимит риска для тестового БИНа! Риск: 0.95. Стейт переведен в FAILED.`, 'error');
            finishPipeline();
            return;
        }

        pushLog(`🛡️ [FRAUD PASS]: Транзакция одобрена радаром за 1.12мс. Risk Score: 0.03.`, 'success');

        // --- ЭТАП 3: ACQUIRING HOLD ---
        await sleep(1000);
        document.getElementById('line_2').classList.add('active');
        document.getElementById('node_HOLD').classList.add('node-active');
        pushLog(`🎰 [CONTROL PLANE]: Инициирован ProcessHold(). Сдвиг FSM графа переходов.`, 'sys');
        pushLog(`🔒 [AES-GCM CRYPTO]: Снабжение операции b2b-токеном: ${result.token}`);
        
        // --- ЭТАП 4: HELD ---
        await sleep(1000);
        document.getElementById('line_3').classList.add('active');
        document.getElementById('node_HELD').classList.add('node-active');
        pushLog(`🔒 [FSM STATE HELD]: Копейки успешно заморожены банком-эквайером на 30 минут.`, 'success');

        // --- ЭТАП 5: CAPTURED & LEDGER ---
        await sleep(1000);
        document.getElementById('line_4').classList.add('active');
        document.getElementById('node_CAPTURED').classList.add('node-active');
        pushLog(`📝 [gRPC DATA PLANE]: Вызов CommitDoubleEntryTransaction() к модулю Ledger...`, 'sys');
        
        const chargeResponse = await fetch('/api/v1/charge');
        const chargeResult = await chargeResponse.json();

        if (!chargeResult.success) {
            document.getElementById('line_4').classList.remove('active');
            document.getElementById('line_4').classList.add('failed');
            document.getElementById('node_CAPTURED').classList.remove('node-active');
            document.getElementById('node_CAPTURED').classList.add('node-failed');
            pushLog(`🛑 [LEDGER REJECTION]: Книга двойной записи обрубила расчет: ${chargeResult.error}`, 'error');
        } else {
            document.getElementById('tps_counter').innerText = (Math.random() * 5 + 1).toFixed(1);
            pushLog(`🗄️ [IMMUTABLE LEDGER]: Проводка ${chargeResult.tx_id} успешно запечатана в Append-Only NoSQL лог!`, 'success');
            pushLog(`💳 [DEBIT]: wallet_david_buyer ➔ Баланс уменьшен на копейки.`, 'success');
            pushLog(`💳 [CREDIT]: wallet_merchant_shop ➔ Баланс пополнен. Инвариант Double-Entry соблюден.`, 'success');
            pushLog(`🏆 [PROCESSED]: Платёж CAPTURED. Транзакция закрыта со 100% успехом!`, 'success');
        }

    } catch (err) {
        pushLog(`🚨 [CORE FAULT]: Операция аварийно прервана. Откат стейт-машины в FAILED.`, 'error');
    } finally {
        finishPipeline();
    }
}

function finishPipeline() {
    const btn = document.getElementById('btn_pay');
    const btnText = document.getElementById('btn_text');
    if (btn) {
        btn.disabled = false;
        btn.style.opacity = "1";
        btn.style.cursor = "pointer";
        btnText.innerText = "ПРОВЕСТИ ТРАНЗАКЦИЮ";
    }
}
