import './Tasks.css';

const Tasks = () => {
    return (
        <div className="tasks-container">
            <div className="tasks">
                <div className="tasks__content">
                    <h1 className="tasks__header">Ежедневные задания</h1>
                    <p className="tasks__description">Выполняй задания, чтобы развивать питомца</p>
                </div>
                <div className="tasks__stats">
                    <h1 className="tasks__stats-title">Задания на сегодня</h1>
                        <div className="tasks__progress">
                            <div className="tasks__progress-bar">
                                <div
                                    className="tasks__progress-fill tasks__progress-fill--xp"
                                    style={{ width: '0%' }}
                                />
                            </div>
                            <div className="tasks__progress-info">
                                <span className="tasks__progress-label">Выполнено: 0 из 3</span>
                            </div>
                        </div>
                </div>
                <div className="tasks__list">
                    <TaskItem/>
                    <TaskItem/>
                    <TaskItem/>
                </div>
            </div>
        </div>
    );
};

const TaskItem = () => {
    return(
        <div className="tasks__card-item">
            <div className="tasks__card-icon">XP</div>
            <div className="tasks__card-info">
                <h1 className="tasks__card-title">Просмотри 3 объявления</h1>
                <p className="tasks__card-description">Посмотри любые три товара в поиске</p>
            </div>
            <button className="tasks__complete-button">Выполнить</button>
        </div>
    )
}

export default Tasks;