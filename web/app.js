/**
 * @param {{id: number, content: string, created: string, children: Array}} comment
 */
function renderComment(comment) {
    const div = document.createElement('div');
    div.className = 'comment-node';
    div.id = `comment-${comment.id}`;

    div.innerHTML = `
        <div class="comment-body">
            <p class="content">${comment.content}</p>
            <div class="actions">
                <small>${new Date(comment.created).toLocaleString()}</small>
                <button class="btn-reply" onclick="showReplyForm(${comment.id})">Ответить</button>
                <button class="btn-delete" onclick="deleteComment(${comment.id})">Удалить</button>
            </div>
            <div id="reply-form-${comment.id}" class="reply-form" style="display:none">
                <textarea id="text-${comment.id}" placeholder="Ваш ответ..."></textarea>
                <button onclick="sendReply(${comment.id})">Отправить</button>
            </div>
        </div>
        <div class="children"></div>
    `;

    if (comment.children && comment.children.length > 0) {
        const childrenContainer = div.querySelector('.children');
        comment.children.forEach(child => {
            childrenContainer.appendChild(renderComment(child));
        });
    }
    return div;
}


function showReplyForm(id) {
    const form = document.getElementById(`reply-form-${id}`);
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
}

async function sendReply(parentId) {
    const content = document.getElementById(`text-${parentId}`).value;
    if (!content) return;

    await fetch('/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ parent_id: parentId, content: content })
    });
    location.reload(); // Перезагружаем для простоты, чтобы увидеть результат
}

async function deleteComment(id) {
    if (!confirm('Удалить этот комментарий и все ответы?')) return;

    await fetch(`/comments/${id}`, { method: 'DELETE' });
    location.reload();
}

async function sendMainComment() {
    const content = document.getElementById('main-content').value;
    if (!content.trim()) return;

    const response = await fetch('/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: content })
    });

    if (response.ok) {
        document.getElementById('main-content').value = ''; // Очищаем поле
        await loadComments(); // Перезагружаем дерево без перезагрузки страницы
    } else {
        const err = await response.json();
        alert("Ошибка: " + err.error);
    }

    location.reload()
}

// Функция поиска (вызывается по кнопке "Найти")
async function searchComments() {
    const query = document.getElementById('search-input').value;
    if (query.length < 3) {
        alert("Введите минимум 3 символа для поиска");
        return;
    }

    try {
        const response = await fetch(`/comments/search?q=${encodeURIComponent(query)}`);
        const results = await response.json();

        const container = document.getElementById('comments-container');
        container.innerHTML = `<h3>Результаты поиска по запросу: "${query}"</h3>`;

        if (results.length === 0) {
            container.innerHTML += '<p>Ничего не найдено :(</p>';
        } else {
            results.forEach(comment => {
                container.appendChild(renderComment(comment));
            });
        }

        // Показываем кнопку сброса, чтобы можно было вернуться к дереву
        document.getElementById('reset-search').style.display = 'inline-block';
    } catch (err) {
        console.error('Ошибка поиска:', err);
    }
}

let currentPage = 1;
const limit = 1; // Количество веток на страницу

// Полный сброс: поиск очищен, страница первая, контейнер пустой
async function loadComments() {
    // Сброс пагинации
    currentPage = 1;
    const loadMoreBtn = document.getElementById('load-more-btn');
    if (loadMoreBtn) loadMoreBtn.style.display = 'block';

    // Сброс поиска
    const searchInput = document.getElementById('search-input');
    if (searchInput) searchInput.value = '';
    const resetBtn = document.getElementById('reset-search');
    if (resetBtn) resetBtn.style.display = 'none';

    // Очистка контейнера
    const container = document.getElementById('comments-container');
    container.innerHTML = '';

    // Загрузка первой страницы
    await fetchAndRender();
}

// Функция подгрузки следующей страницы
async function loadMore() {
    currentPage++;
    await fetchAndRender();
}

// Единая функция для запроса данных
async function fetchAndRender() {
    try {
        const response = await fetch(`/comments?page=${currentPage}&limit=${limit}`);
        const trees = await response.json();

        const container = document.getElementById('comments-container');
        const loadMoreBtn = document.getElementById('load-more-btn');

        if (trees.length === 0) {
            if (loadMoreBtn) loadMoreBtn.style.display = 'none';
            if (currentPage === 1) container.innerHTML = '<p>Комментариев пока нет</p>';
            return;
        }

        // Здесь мы используем appendChild, чтобы ДОБАВЛЯТЬ к списку, а не заменять его
        trees.forEach(tree => {
            container.appendChild(renderComment(tree));
        });

        if (trees.length < limit) {
            if (loadMoreBtn) loadMoreBtn.style.display = 'none';
        } else {
            // Если пришло РОВНО limit, данные МОГУТ быть еще,
            // показываем кнопку для следующей страницы.
            if (loadMoreBtn) loadMoreBtn.style.display = 'block';
        }
    } catch (err) {
        console.error('Ошибка загрузки данных:', err);
    }
}

loadComments().catch(console.error);


