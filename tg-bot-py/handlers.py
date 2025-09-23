import logging
import requests
from telegram import Update, InlineKeyboardMarkup, InlineKeyboardButton
from telegram.ext import ContextTypes, ConversationHandler
from dataclasses import dataclass, field
from typing import Dict, Optional
from config import config

API_URL = config['bot']['api_url']

logger = logging.getLogger(__name__)

@dataclass
class UserState:
    captcha_id: Optional[str] = None
    action: Optional[str] = None
    task_data: Dict = field(default_factory=dict)
    offer_data: Dict = field(default_factory=dict)
#STATES
user_states: Dict[int, UserState] = {}
#JWT
sessions: Dict[int, str] = {}

# Conversation states
ASK_ADDRESS, ASK_AMOUNT = range(2)
ASK_TASK_TITLE, ASK_TASK_DESC, ASK_TASK_PRICE, ASK_TASK_DEADLINE = range(4)

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    user_id = update.message.from_user.id
    if user_id in sessions:
        await update.message.reply_text("✅ Вы уже авторизованы.")
        await user_menu(update, context)
        return
    keyboard = [
        [InlineKeyboardButton("Зарегистрироваться", callback_data="register")],
        [InlineKeyboardButton("Войти", callback_data="auth")],
        [InlineKeyboardButton("Восстановить", callback_data="restore")]
    ]
    await update.message.reply_text(
        "Привет! Выберите действие:",
        reply_markup=InlineKeyboardMarkup(keyboard)
    )

# Universal handler of buttons
async def action_button(update: Update, context: ContextTypes.DEFAULT_TYPE):
    query = update.callback_query
    await query.answer()

    action = query.data  # register / auth / restore

    try:
        resp = requests.get(f"{API_URL}/captcha", headers={"Accept": "image/png"})
    except Exception as e:
        logger.error(f"Ошибка подключения: {e}")
        resp = None

    captcha_id = None
    if resp and resp.status_code == 200:
        captcha_id = resp.headers.get("X-Captcha-ID")

    # сохраняем состояние
    user_states[query.from_user.id] = UserState(captcha_id=captcha_id, action=action)

    if captcha_id:  # капча включена
        if action == "register":
            text = "Введите: <капча> <логин> <пароль>"
        elif action == "auth":
            text = "Введите: <капча> <логин> <пароль>"
        elif action == "restore":
            text = "Введите: <капча> <логин> <мнемоника> <новый_пароль>"
        await query.message.reply_photo(resp.content, caption=text)
    else:  # капча выключена
        if action == "register":
            text = "Введите: <логин> <пароль>"
        elif action == "auth":
            text = "Введите: <логин> <пароль>"
        elif action == "restore":
            text = "Введите: <логин> <мнемоника> <новый_пароль>"
        await query.message.reply_text(text)

async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    user_id = update.message.from_user.id
    if user_id not in user_states:
        await update.message.reply_text("Сначала выберите действие через /start.")
        return

    state = user_states[user_id]
    captcha_id = state.captcha_id
    action = state.action

    parts = update.message.text.split(" ")
    data = {}
    url = None

    # ---- REGISTER ----
    if action == "register":
        if captcha_id:
            if len(parts) < 3:
                await update.message.reply_text("Формат: <капча> <логин> <пароль>")
                return
            captcha_text, login, password = parts[0], parts[1], " ".join(parts[2:])
            data = {"captcha_id": captcha_id, "captcha_answer": captcha_text,
                    "username": login, "password": password}
        else:
            if len(parts) < 2:
                await update.message.reply_text("Формат: <логин> <пароль>")
                return
            login, password = parts[0], " ".join(parts[1:])
            data = {"username": login, "password": password}
        url = f"{API_URL}/register"

    # ---- AUTH ----
    elif action == "auth":
        if captcha_id:
            if len(parts) < 3:
                await update.message.reply_text("Формат: <капча> <логин> <пароль>")
                return
            captcha_text, login, password = parts[0], parts[1], " ".join(parts[2:])
            data = {"captcha_id": captcha_id, "captcha_answer": captcha_text,
                    "username": login, "password": password}
        else:
            if len(parts) < 2:
                await update.message.reply_text("Формат: <логин> <пароль>")
                return
            login, password = parts[0], " ".join(parts[1:])
            data = {"username": login, "password": password}
        url = f"{API_URL}/auth"

    # ---- RESTORE ----
    elif action == "restore":
        if captcha_id:
            if len(parts) < 4:
                await update.message.reply_text("Формат: <капча> <логин> <мнемоника> <новый_пароль>")
                return
            captcha_text = parts[0]
            login = parts[1]
            new_pass = parts[-1]  # последний элемент
            mnemonic = " ".join(parts[2:-1])
            data = {
                "captcha_id": captcha_id,
                "captcha_answer": captcha_text,
                "username": login,
                "mnemonic": mnemonic,
                "new_password": new_pass
            }
        else:
            if len(parts) < 3:
                await update.message.reply_text("Формат: <логин> <мнемоника> <новый_пароль>")
                return
            login = parts[0]
            new_pass = parts[-1]
            mnemonic = " ".join(parts[1:-1])
            data = {
                "username": login,
                "mnemonic": mnemonic,
                "new_password": new_pass
            }
        url = f"{API_URL}/restoreuser"

    try:
        resp = requests.post(url, json=data)
    except Exception as e:
        await update.message.reply_text(f"Ошибка подключения: {e}")
        return

    if resp.status_code == 200:
        try:
            js = resp.json()
        except Exception:
            js = {"message": resp.text}

        token = js.get("token") or js.get("encrypted")
        if token:
            sessions[user_id] = token
            await update.message.reply_text(f"{js.get('message', 'OK')}\nТокен сохранён.")
            await user_menu(update, context)
        else:
            await update.message.reply_text(js.get("message", resp.text))
    else:
        await update.message.reply_text(f"Ошибка {resp.status_code}: {resp.text}")

    user_states.pop(user_id, None)

async def menu_button(update: Update, context: ContextTypes.DEFAULT_TYPE):
    query = update.callback_query
    await query.answer()

    user_id = query.from_user.id
    if user_id not in sessions:
        await query.message.reply_text("⚠️ Вы не авторизованы. Используйте /start.")
        return

    token = sessions[user_id]
    action = query.data

    if action == "menu_tasks":
        keyboard = [
            [InlineKeyboardButton("Создать задание", callback_data="task_create")],
            [InlineKeyboardButton("Мои задания", callback_data="task_list")],
            [InlineKeyboardButton("◀️ Назад", callback_data="back_start")]
        ]
        await query.message.reply_text("Задания:", reply_markup=InlineKeyboardMarkup(keyboard))

    elif action == "task_list":
        try:
            resp = requests.get(
                f"{API_URL}/api/tasks",
                headers={"Authorization": f"Bearer {token}"}
            )
        except Exception as e:
            await query.message.reply_text(f"Ошибка подключения: {e}")
            return

        if resp.status_code == 200:
            js = resp.json()
            tasks = js.get("tasks", [])
            if not tasks:
                await query.message.reply_text("У вас нет заданий.")
            else:
                text = "Ваши задания:\n"
                for task in tasks[:5]:  # limit to 5
                    text += f"- {task['title']} (ID: {task['id']})\n"
                await query.message.reply_text(text)
        else:
            await query.message.reply_text(f"Ошибка {resp.status_code}: {resp.text}")

    elif action == "menu_wallet":
        keyboard = [
            [InlineKeyboardButton("📤 Отправить BTC", callback_data="wallet_send")],
            [InlineKeyboardButton("◀️ Назад", callback_data="back_start")]
        ]
        try:
            resp = requests.get(
                f"{API_URL}/api/wallet",
                params={"currency": "BTC"},
                headers={"Authorization": f"Bearer {token}"}
            )
        except Exception as e:
            await query.message.reply_text(f"Ошибка подключения к API: {e}")
            return

        if resp.status_code == 200:
            try:
                js = resp.json()
            except Exception:
                await query.message.reply_text("Ошибка разбора ответа от API")
                return

            address = js.get("address", "неизвестно")
            balance = js.get("balance", 0)
            await query.message.reply_text(
                f"💰 Ваш BTC-кошелёк:\n\n"
                f"📍 Адрес: `{address}`\n"
                f"💵 Баланс: {balance} BTC",
                parse_mode="Markdown", reply_markup=InlineKeyboardMarkup(keyboard)
            )
        else:
            await query.message.reply_text(f"Ошибка {resp.status_code}: {resp.text}")

    elif action == "back_start":
        await start(update, context)
    else:
        await query.message.reply_text("❓ Функция пока не реализована.")


async def ask_address(update: Update, context: ContextTypes.DEFAULT_TYPE):
    context.user_data["btc_address"] = update.message.text.strip()
    await update.message.reply_text("Введите сумму в BTC:")
    return ASK_AMOUNT

async def ask_amount(update: Update, context: ContextTypes.DEFAULT_TYPE):
    amount = update.message.text.strip()
    address = context.user_data.get("btc_address")
    token = context.user_data.get("token")

    try:
        resp = requests.post(
            f"{API_URL}/api/wallet/bitcoinSend",
            params={"to": address, "amount": amount},
            headers={"Authorization": f"Bearer {token}"}
        )
    except Exception as e:
        await update.message.reply_text(f"Ошибка подключения: {e}")
        return ConversationHandler.END

    if resp.status_code == 200:
        try:
            js = resp.json()
        except Exception:
            js = {"message": resp.text}

        await update.message.reply_text(
            f"✅ Транзакция отправлена\n\n"
            f"📍 Адрес: {js.get('to', address)}\n"
            f"💵 Сумма: {js.get('remaining', amount)} BTC\n"
            f"💸 Комиссия: {js.get('commission', '0')} BTC"
        )
    else:
        await update.message.reply_text(f"Ошибка {resp.status_code}: {resp.text}")

    return ConversationHandler.END

async def task_create_entry(update: Update, context: ContextTypes.DEFAULT_TYPE):
    query = update.callback_query
    await query.answer()
    user_id = query.from_user.id
    if user_id not in sessions:
        await query.message.reply_text("⚠️ Вы не авторизованы.")
        return ConversationHandler.END
    token = sessions[user_id]
    context.user_data["token"] = token
    await query.message.reply_text("Введите название задания:")
    return ASK_TASK_TITLE

async def ask_task_title(update: Update, context: ContextTypes.DEFAULT_TYPE):
    context.user_data["task_title"] = update.message.text.strip()
    await update.message.reply_text("Введите описание задания:")
    return ASK_TASK_DESC

async def ask_task_desc(update: Update, context: ContextTypes.DEFAULT_TYPE):
    context.user_data["task_desc"] = update.message.text.strip()
    await update.message.reply_text("Введите цену в BTC:")
    return ASK_TASK_PRICE

async def ask_task_price(update: Update, context: ContextTypes.DEFAULT_TYPE):
    context.user_data["task_price"] = update.message.text.strip()
    await update.message.reply_text("Введите дедлайн (YYYY-MM-DDTHH:MM:SSZ):")
    return ASK_TASK_DEADLINE

async def ask_task_deadline(update: Update, context: ContextTypes.DEFAULT_TYPE):
    deadline = update.message.text.strip()
    title = context.user_data.get("task_title")
    desc = context.user_data.get("task_desc")
    price = context.user_data.get("task_price")
    token = context.user_data.get("token")

    data = {
        "title": title,
        "description": desc,
        "price": float(price),
        "currency": "BTC",
        "deadline": deadline
    }

    try:
        resp = requests.post(
            f"{API_URL}/api/tasks",
            json=data,
            headers={"Authorization": f"Bearer {token}"}
        )
    except Exception as e:
        await update.message.reply_text(f"Ошибка подключения: {e}")
        return ConversationHandler.END

    if resp.status_code == 200:
        js = resp.json()
        await update.message.reply_text(f"Задание создано: {js.get('task', {}).get('title')}")
    else:
        await update.message.reply_text(f"Ошибка {resp.status_code}: {resp.text}")

    return ConversationHandler.END

async def user_menu(update: Update, context: ContextTypes.DEFAULT_TYPE):
    keyboard = [
        [InlineKeyboardButton("Задания", callback_data="menu_tasks")],
        [InlineKeyboardButton("Кошелек", callback_data="menu_wallet")],
        [InlineKeyboardButton("Отзывы", callback_data="menu_reviews")],
        [InlineKeyboardButton("Диспуты", callback_data="menu_disputes")],
        [InlineKeyboardButton("Профиль", callback_data="menu_profile")],
        [InlineKeyboardButton("Тикеты", callback_data="menu_tickets")],
        [InlineKeyboardButton("Чаты", callback_data="menu_chats")],
        [InlineKeyboardButton("Назад", callback_data="back_start")]
    ]
    if update.message:
        await update.message.reply_text("📋 Меню пользователя:", reply_markup=InlineKeyboardMarkup(keyboard))
    else:
        await update.callback_query.message.reply_text("📋 Меню пользователя:", reply_markup=InlineKeyboardMarkup(keyboard))
