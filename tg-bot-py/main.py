import logging
from telegram.ext import (
    Application, CommandHandler, CallbackQueryHandler,
    MessageHandler, filters, ConversationHandler
)
from config import config
from handlers import (
    start, action_button, handle_message, menu_button,
    ask_address, ask_amount, task_create_entry,
    ask_task_title, ask_task_desc, ask_task_price, ask_task_deadline,
    ASK_ADDRESS, ASK_AMOUNT, ASK_TASK_TITLE, ASK_TASK_DESC, ASK_TASK_PRICE, ASK_TASK_DEADLINE
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def main():
    # TODO: бот пока не работает, это шаблон
    app = Application.builder().token(config['bot']['token']).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CallbackQueryHandler(action_button, pattern="register|auth|restore"))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))
    app.add_handler(CallbackQueryHandler(menu_button, pattern="menu_.*|back_start|task_.*"))

    # Wallet send conversation
    wallet_conv = ConversationHandler(
        entry_points=[CallbackQueryHandler(menu_button, pattern="wallet_send")],
        states={
            ASK_ADDRESS: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_address)],
            ASK_AMOUNT: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_amount)],
        },
        fallbacks=[],
    )

    # Task create conversation
    task_conv = ConversationHandler(
        entry_points=[CallbackQueryHandler(task_create_entry, pattern="task_create")],
        states={
            ASK_TASK_TITLE: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_task_title)],
            ASK_TASK_DESC: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_task_desc)],
            ASK_TASK_PRICE: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_task_price)],
            ASK_TASK_DEADLINE: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_task_deadline)],
        },
        fallbacks=[],
    )

    app.add_handler(wallet_conv)
    app.add_handler(task_conv)
    app.run_polling()

if __name__ == "__main__":
    main()

