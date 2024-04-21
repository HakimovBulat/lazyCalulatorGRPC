# lazyCalulatorGRPC
Данный веб-сайт предназначен для вычисления простейших выражений: сложение, вычитание, умножение и деление. Но с одной оговоркой - каждое такое действие занимает немалое количество времени, которое пользователь может редактировать. <br>
Для редакции времени операторов - надо зарегистрироваться, а потом войти по логину <br>
В данном проекте использовалась одна главная СУБД - Postgresql и БД: Expression и Users.<br>
Для грамотной работы установите Postgresql и в router/router.go поменяйте в строке connection пароль и имя пользователя на свои <br>
Возможности сайта:<br>
/ - вычисление элементарных выражений путем POST-запроса в единственной форме. Снизу будут видны все выражения, вычисленные до настоящего времени.Чтобы отправить выражение - нажмите Enter. Кнопка "Обновить данные" - обновляет историю поиска.<br>
/static_operators - показ вычислительных мощностей на каждом конкретном действии. Для редакции - нажать на единственную кнопку.<br>
/operators - редактирование вычислительных мощностей. Переход на эту ссылку произведен с помощью HTMX (для меня выйчить JavaScript было нереально).<br>
/get_expression/:id - узнать информацию о конкретном примере по его Id. Реализован через JSON.<br>
/register - регистрация пользователя<br>
/login - вход по логину <br>
/logout - выход из профиля (надо нажать на свой логин на главной странице)
телеграмм: @BulatHakimov

СХЕМА
===

ПОЛЬЗОВАТЕЛЬ <----- ------> WebServer <----- ------> База Данных<br>
PS
===

Да проект называется lazyCalulatorGRPC, но gRPC не успел реализовать. Но хотя бы SQL. Надеюсь на поинмание и заранее спасибо