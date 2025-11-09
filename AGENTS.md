Отвечай на русском языке.

After receiving tool results, carefully reflect on their quality and determine optimal next steps before proceeding. Use your thinking to plan and iterate based on this new information, and then take the best next action.
For maximum efficiency, whenever you need to perform multiple independent operations, invoke all relevant tools simultaneously rather than sequentially.
If you create any temporary new files, scripts, or helper files for iteration, clean up these files by removing them at the end of the task.
Please write a high quality, general purpose solution. Implement a solution that works correctly for all valid inputs, not just the test cases. Do not hard-code values or create solutions that only work for specific test inputs. Instead, implement the actual logic that solves the problem generally.
Focus on understanding the problem requirements and implementing the correct algorithm. Tests are there to verify correctness, not to define the solution. Provide a principled implementation that follows best practices and software design principles.
If the task is unreasonable or infeasible, or if any of the tests are incorrect, please tell me. The solution should be robust, maintainable, and extendable.

The existing code structure must not be changed without a strong reason.
Every bug must be reproduced by a unit test before being fixed.
Every new feature must be covered by a unit test before it is implemented.
Every newly introduced method must have a supplementary docblock preceding it.
Minor inconsistencies and typos in the existing code may be fixed.
Method and function bodies may not contain comments.
Favor "fail fast" paradigm over "fail safe": throw exception earlier.

Constructors may not contain any code except assignment statements.
Every class may encapsulate no more than four attributes.
Every class must encapsulate at least one attribute.
Getters must be avoided, as they are symptoms of an anemic object model.
Implementation inheritance must be avoided at all costs (not to be confused with subtyping).
The DDD paradigm must be respected.
Setters must be avoided, as they make objects mutable.
Immutable objects must be favored over mutable ones.
Static methods in classes are strictly prohibited.
Method names must respect the CQRS principle: they must be either nouns or verbs.
Methods must be declared in interfaces and then implemented in classes.
Public methods that do not implement an interface must be avoided.
Methods must never return null.
Exception messages must include as much context as possible.
Every class must have a supplementary docblock preceding it.
A class docblock must explain the purpose of the class and provide usage examples.
Every method and function must have a supplementary docblock preceding it.
Objects must not provide functionality used only by tests.
The testing framework must be configured to disable logging from the objects under test.

Все комментарии в коде - на английском языке, запрещено использование кириллицы.

- Форматирование и стиль
  - Обязателен линтинг: `ruff check` для `scripts/` и `simple_ctr/`.
  - Форматирование: `ruff format --check`.
  - Сортировка импортов обязательна (правила `I`).
  - Актуализация синтаксиса (правила `UP`), Bugbear (`B`), PIE, Ruff-specific (`RUF`, исключая `RUF001,RUF002,RUF003,RUF012`).

- Дата и время
  - Использовать только timezone-aware `datetime` (правила `DTZ`), например `datetime.now(timezone.utc)`.

- Безопасность и производительность
  - Проверки безопасности (`S`, игнор `S101,S603`).
  - Проверки производительности (`PERF`, игнор `PERF203`).

- Сложность и упрощение кода
  - Когнитивная сложность любой функции ≤ 15.
  - Упрощать конструкции согласно правилам `SIM` (игнор `SIM108`).

- Сигнатуры функций и аргументы
  - Неиспользуемые аргументы запрещены; если необходимо — префикс `_` (правила `ARG`).
  - Не более 5 аргументов у функций, из них не более 3 позиционных (конфиг `PLR0917`).

- Классы
  - Не более 20 публичных методов в классе (конфиг `PLR0904`).

- Документация
  - Обязательные докстринги для публичных модулей, классов, функций и методов (правила `D100-D103`).

- Исключения
  - Сообщения исключений через интерполяцию (`EM101-EM103`).
  - Запрещены «голые» `except` (правило `BLE001`).
  - Корректное связывание исключений `raise ... from ...` там, где применимо (`B904`).
  - Соблюдать лучшие практики `try/except/finally` (правила `TRY002,TRY003,TRY200,TRY201,TRY300,TRY400`).

- Аннотации типов
  - Выполнять правила `ANN001,ANN002,ANN003,ANN201,ANN202,ANN204,ANN205,ANN206`.
  - Статическая проверка типов `pyright`: 0 ошибок (варнинги допустимы).

- Логирование
  - Следовать `LOG001,LOG007,LOG009` (корректная передача параметров без f-строк и т.д.).

- Возвраты и булева логика
  - Правила по возвратам/ветвлениям (`RET`) — единообразные пути возврата.
  - Избегать анти‑паттернов с булевыми аргументами/сравнениями (`PLC1901`).

- Глобальные и прочие предупреждения
  - Устранять предупреждения `PLW0211,PLW0602,PLW0603,PLW0604,PLW1641` (глобалы, мутации, парность `__eq__/__hash__`, и т.п.).

- Доступ к приватным членам
  - Запрещён доступ к приватным атрибутам извне (правило `SLF001`).

- Мёртвый код
  - Запрещён мёртвый код по `vulture` (минимальная уверенность 60; использовать `.vulture_whitelist` при необходимости).

- Доступ к атрибутам
  - Запрещены `hasattr` и `getattr`. Альтернативы: `Protocol/ABC`, `try/except AttributeError`, прямой доступ при гарантированном контракте.

- Тесты
  - Лимит на длину тест‑функций в текущей конфигурации отключён.