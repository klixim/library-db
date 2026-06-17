document.addEventListener('DOMContentLoaded', () => {
    // Утилита для запросов с токеном
    const fetchWithAuth = async (url, options = {}) => {
        const token = localStorage.getItem('token');
        if (token) {
            options.headers = { ...options.headers, 'Authorization': `Bearer ${token}` };
        }
        return fetch(url, options);
    };

    // --- Логика авторизации (Глобальная) ---
    const loginForm = document.getElementById('loginForm');
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    const userInfo = document.getElementById('userInfo');
    const userRoleSpan = document.getElementById('userRole');
    const loginError = document.getElementById('loginError');

    const updateAuthUI = () => {
        const token = localStorage.getItem('token');
        const role = localStorage.getItem('role');
        
        document.querySelectorAll('.admin-only, .reader-only').forEach(el => el.style.display = 'none');

        if (token && role) {
            loginBtn.classList.add('d-none');
            userInfo.classList.remove('d-none');
            userRoleSpan.textContent = role === 'admin' ? 'Администратор' : 'Читатель';

            if (role === 'admin') {
                document.querySelectorAll('.admin-only').forEach(el => el.style.display = 'block');
            } else if (role === 'reader') {
                document.querySelectorAll('.reader-only').forEach(el => el.style.display = 'block');
            }
        } else {
            loginBtn.classList.remove('d-none');
            userInfo.classList.add('d-none');
            
            // Защита страниц от неавторизованного доступа на клиенте
            const path = window.location.pathname;
            if (path === '/profile.html' || path === '/admin.html') {
                window.location.href = '/';
            }
        }
    };

    updateAuthUI();

    if (loginForm) {
        loginForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;

            fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            })
            .then(res => res.json().then(data => ({ status: res.status, body: data })))
            .then(res => {
                if (res.status === 200) {
                    localStorage.setItem('token', res.body.token);
                    localStorage.setItem('role', res.body.role);
                    loginError.classList.add('d-none');
                    
                    const modalEl = document.getElementById('loginModal');
                    const modal = bootstrap.Modal.getInstance(modalEl) || new bootstrap.Modal(modalEl);
                    modal.hide();
                    loginForm.reset();
                    updateAuthUI();
                    if (window.location.pathname === '/book.html') {
                        window.location.reload(); // Перезагружаем страницу, чтобы подтянулась защищенная история
                    }
                } else {
                    loginError.classList.remove('d-none');
                    loginError.textContent = res.body.error || 'Ошибка входа';
                }
            }).catch(err => console.error('Ошибка входа:', err));
        });
    }

    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => {
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            updateAuthUI();
        });
    }

    // Глобальная навигация на страницу книги с параметрами
    window.openBookPage = (id, title) => {
        window.location.href = `/book.html?id=${id}&title=${encodeURIComponent(title)}`;
    };

    // --- Логика для конкретных страниц (Роутинг на клиенте) ---
    
    // 1. Главная страница (Каталог)
    const booksTableBody = document.getElementById('books-table-body');
    if (booksTableBody) {
        const loadCatalog = (query = "") => {
            const url = query ? `/api/books/search?q=${encodeURIComponent(query)}` : '/api/books';
            fetch(url).then(res => res.json()).then(data => {
                if (!data || data.length === 0) return booksTableBody.innerHTML = '<tr><td colspan="5" class="text-center">Ничего не найдено</td></tr>';
                booksTableBody.innerHTML = data.map(book => {
                    const genresHtml = book.genres ? book.genres.split(', ').map(g => `<span class="badge bg-info text-dark me-1">${g}</span>`).join('') : '<span class="text-muted">—</span>';
                    return `<tr>
                        <td>${book.id}</td>
                        <td><strong>${book.title}</strong><br><small class="text-muted">${book.authors}</small></td>
                        <td>${genresHtml}</td>
                        <td>${book.isbn}</td>
                        <td>${book.year}</td>
                        <td><button class="btn btn-sm btn-primary" onclick="openBookPage(${book.id}, '${book.title}')">Подробнее</button></td>
                    </tr>`;
                }).join('');
            }).catch(err => booksTableBody.innerHTML = '<tr><td colspan="5" class="text-center text-danger">Ошибка сети</td></tr>');
        };

        document.getElementById('searchBtn').addEventListener('click', () => {
            loadCatalog(document.getElementById('searchInput').value);
        });

        loadCatalog();
    }

    // 2. Страница книги
    const bpTitle = document.getElementById('bp-title');
    if (bpTitle) {
        const params = new URLSearchParams(window.location.search);
        const id = params.get('id');
        if (id) {
            document.getElementById('bp-id').textContent = id;
            bpTitle.textContent = params.get('title') || 'Книга';
            
            const statusEl = document.getElementById('bp-status');
            fetch(`/api/books/${id}/status`).then(res => res.json()).then(data => {
                statusEl.innerHTML = data.is_available 
                    ? '<span class="badge bg-success fs-6">Доступна в библиотеке</span>' 
                    : '<span class="badge bg-secondary fs-6">На руках</span>';
            }).catch(() => statusEl.innerHTML = '<span class="text-danger">Ошибка</span>');

            // Загрузка истории выдач книги (только для админа)
            const role = localStorage.getItem('role');
            if (role === 'admin') {
                fetchWithAuth(`/api/admin/books/${id}/history`)
                    .then(res => res.json())
                    .then(data => {
                        if (data.error) {
                            document.getElementById('book-history-body').innerHTML = `<tr><td colspan="4" class="text-center text-danger">Ошибка сервера: ${data.error}</td></tr>`;
                            return;
                        }
                        const tbody = document.getElementById('book-history-body');
                        if (!data || data.length === 0) return tbody.innerHTML = '<tr><td colspan="4" class="text-center">История пуста</td></tr>';
                        tbody.innerHTML = data.map(h => `<tr>
                            <td>${h.loan_id}</td>
                            <td>${h.reader_name}</td>
                            <td>${new Date(h.loan_date).toLocaleDateString()}</td>
                            <td>${h.return_date ? new Date(h.return_date).toLocaleDateString() : '<span class="text-warning fw-bold">На руках</span>'}</td>
                        </tr>`).join('');
                    }).catch(() => {
                        document.getElementById('book-history-body').innerHTML = '<tr><td colspan="4" class="text-center text-danger">Ошибка доступа</td></tr>';
                    });

                const exportBtn = document.getElementById('exportHistoryBtn');
                if (exportBtn) {
                    exportBtn.addEventListener('click', () => {
                        const token = localStorage.getItem('token');
                        fetch(`/api/admin/books/${id}/export-history`, {
                            headers: { 'Authorization': `Bearer ${token}` }
                        })
                        .then(res => res.blob())
                        .then(blob => {
                            const url = window.URL.createObjectURL(blob);
                            const a = document.createElement('a');
                            a.href = url;
                            a.download = `book_${id}_history.csv`;
                            document.body.appendChild(a);
                            a.click();
                            a.remove();
                        });
                    });
                }
            }
        }
    }

    // 3. Личный кабинет (Запросы 3 и 4)
    const currentLoansList = document.getElementById('current-loans-list');
    if (currentLoansList) {
        const readerId = 1; 
        
        fetchWithAuth(`/api/readers/${readerId}/loans`).then(r => r.json()).then(data => {
            if (!data || data.length === 0) return currentLoansList.innerHTML = '<li class="list-group-item text-muted">У вас нет книг на руках</li>';
            currentLoansList.innerHTML = data.map(l => `<li class="list-group-item d-flex justify-content-between align-items-center">
                ${l.title} 
                <span class="badge bg-warning text-dark rounded-pill">На руках: ${l.days_on_loan} дн.</span>
            </li>`).join('');
        });

        fetchWithAuth(`/api/readers/${readerId}/history`).then(r => r.json()).then(data => {
            const list = document.getElementById('history-loans-list');
            if (!data || data.length === 0) return list.innerHTML = '<li class="list-group-item text-muted">История пуста</li>';
            list.innerHTML = data.map(l => `<li class="list-group-item">
                ${l.title} <span class="text-muted float-end">Возвращена: ${new Date(l.return_date).toLocaleDateString()}</span>
            </li>`).join('');
        });
    }

    // 4. Админ-панель (Запрос 5 и 6)
    const debtorsTableBody = document.getElementById('debtors-table-body');
    if (debtorsTableBody) {
        fetchWithAuth('/api/admin/debtors')
            .then(res => res.json())
            .then(data => {
                if (!data || data.length === 0) return debtorsTableBody.innerHTML = '<tr><td colspan="4" class="text-center">Должников нет!</td></tr>';
                debtorsTableBody.innerHTML = data.map(d => `<tr>
                    <td>${d.full_name}</td>
                    <td>${d.email} <br> <small>${d.phone || 'Нет телефона'}</small></td>
                    <td>${d.book_title}</td>
                    <td><strong>${d.overdue_days}</strong></td>
                </tr>`).join('');
            }).catch(err => debtorsTableBody.innerHTML = `<tr><td colspan="4" class="text-danger text-center">Ошибка доступа</td></tr>`);

        fetchWithAuth('/api/admin/popular-books')
            .then(res => res.json())
            .then(data => {
                const tbody = document.getElementById('stats-table-body');
                if (!data || data.length === 0) return tbody.innerHTML = '<tr><td colspan="3" class="text-center">Нет статистики</td></tr>';
                tbody.innerHTML = data.map((b, index) => `<tr>
                    <td>${b.id}</td>
                    <td>
                        ${index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉'} 
                        ${b.title}
                    </td>
                    <td>${b.borrow_count}</td>
                </tr>`).join('');
            }).catch(err => document.getElementById('stats-table-body').innerHTML = `<tr><td colspan="3" class="text-danger text-center">Ошибка доступа</td></tr>`);

        // Загрузка списка читателей
        const readersTableBody = document.getElementById('readers-table-body');
        const loadReaders = () => {
            if (!readersTableBody) return;
            fetchWithAuth('/api/admin/readers')
                .then(res => res.json())
                .then(data => {
                    if (!data || data.length === 0) return readersTableBody.innerHTML = '<tr><td colspan="4" class="text-center">Нет читателей</td></tr>';
                    readersTableBody.innerHTML = data.map(r => `<tr>
                        <td>${r.id}</td>
                        <td>${r.full_name}</td>
                        <td>${r.email || '—'}</td>
                        <td>${r.phone || '—'}</td>
                    </tr>`).join('');
                }).catch(err => readersTableBody.innerHTML = `<tr><td colspan="4" class="text-danger text-center">Ошибка сети</td></tr>`);
        };
        loadReaders();

        // Формы выдачи и возврата
        const issueForm = document.getElementById('issueForm');
        if (issueForm) {
            issueForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const readerId = document.getElementById('readerIdIssue').value;
                const bookId = document.getElementById('bookIdIssue').value;
                const msgEl = document.getElementById('issueMessage');
                msgEl.innerHTML = '<span class="text-muted">Отправка...</span>';

                fetchWithAuth('/api/admin/issue', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ reader_id: parseInt(readerId), book_id: parseInt(bookId) })
                })
                .then(res => res.json().then(data => ({ status: res.status, body: data })))
                .then(res => {
                    if (res.status === 200 || res.status === 201) {
                        msgEl.innerHTML = `<span class="text-success">Книга успешно выдана!</span>`;
                        issueForm.reset();
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error || 'Ошибка выдачи'}</span>`;
                    }
                }).catch(err => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }

        const returnForm = document.getElementById('returnForm');
        if (returnForm) {
            returnForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const bookId = document.getElementById('bookIdReturn').value;
                const msgEl = document.getElementById('returnMessage');
                msgEl.innerHTML = '<span class="text-muted">Отправка...</span>';

                fetchWithAuth('/api/admin/return', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ book_id: parseInt(bookId) })
                })
                .then(res => res.json().then(data => ({ status: res.status, body: data })))
                .then(res => {
                    if (res.status === 200) {
                        msgEl.innerHTML = `<span class="text-success">Возврат успешно оформлен!</span>`;
                        returnForm.reset();
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error || 'Ошибка возврата'}</span>`;
                    }
                }).catch(err => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }

        // Добавление книги
        const addBookForm = document.getElementById('addBookForm');
        if (addBookForm) {
            addBookForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const msgEl = document.getElementById('addBookMessage');
                const data = {
                    title: document.getElementById('bookTitle').value,
                    isbn: document.getElementById('bookIsbn').value,
                    publication_year: parseInt(document.getElementById('bookYear').value)
                };
                
                fetchWithAuth('/api/admin/books', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                })
                .then(res => res.json().then(d => ({ status: res.status, body: d })))
                .then(res => {
                    if (res.status === 201) {
                        msgEl.innerHTML = `<span class="text-success">Книга добавлена! Новый ID: ${res.body.id}</span>`;
                        addBookForm.reset();
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error}</span>`;
                    }
                }).catch(() => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }

        // Удаление книги
        const deleteBookForm = document.getElementById('deleteBookForm');
        if (deleteBookForm) {
            deleteBookForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const msgEl = document.getElementById('deleteBookMessage');
                const id = document.getElementById('bookIdDelete').value;
                
                fetchWithAuth(`/api/admin/books/${id}`, { method: 'DELETE' })
                .then(res => res.json().then(d => ({ status: res.status, body: d })))
                .then(res => {
                    if (res.status === 200) {
                        msgEl.innerHTML = `<span class="text-success">Книга успешно удалена!</span>`;
                        deleteBookForm.reset();
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error}</span>`;
                    }
                }).catch(() => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }

        // Добавление читателя
        const addReaderForm = document.getElementById('addReaderForm');
        if (addReaderForm) {
            addReaderForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const msgEl = document.getElementById('addReaderMessage');
                const data = {
                    full_name: document.getElementById('readerName').value,
                    email: document.getElementById('readerEmail').value,
                    phone: document.getElementById('readerPhone').value
                };
                fetchWithAuth('/api/admin/readers', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }).then(res => res.json().then(d => ({ status: res.status, body: d }))).then(res => {
                    if (res.status === 201) {
                        msgEl.innerHTML = `<span class="text-success">Читатель добавлен! Новый ID: ${res.body.id}</span>`;
                        addReaderForm.reset();
                        loadReaders(); // Обновляем таблицу
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error}</span>`;
                    }
                }).catch(() => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }

        // Удаление читателя
        const deleteReaderForm = document.getElementById('deleteReaderForm');
        if (deleteReaderForm) {
            deleteReaderForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const msgEl = document.getElementById('deleteReaderMessage');
                const id = document.getElementById('readerIdDelete').value;
                fetchWithAuth(`/api/admin/readers/${id}`, { method: 'DELETE' })
                .then(res => res.json().then(d => ({ status: res.status, body: d }))).then(res => {
                    if (res.status === 200) {
                        msgEl.innerHTML = `<span class="text-success">Читатель удален!</span>`;
                        deleteReaderForm.reset();
                        loadReaders(); // Обновляем таблицу
                    } else {
                        msgEl.innerHTML = `<span class="text-danger">${res.body.error}</span>`;
                    }
                }).catch(() => msgEl.innerHTML = `<span class="text-danger">Ошибка сети</span>`);
            });
        }
    }
});