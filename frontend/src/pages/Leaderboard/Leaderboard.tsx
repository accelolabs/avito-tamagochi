import './Leaderboard.css';

const players = [
  { place: 1, name: 'Александр К.', xp: 12450 },
  { place: 2, name: 'Елена М.', xp: 11200 },
  { place: 3, name: 'Дмитрий С.', xp: 10850 },
  { place: 4, name: 'Анна В.', xp: 9800 },
  { place: 5, name: 'Иван Иванов', xp: 9300},
  { place: 6, name: 'Сергей П.', xp: 8900 },
  { place: 7, name: 'Ольга Н.', xp: 8450 },
  { place: 8, name: 'Максим Р.', xp: 7900 },
  { place: 9, name: 'Екатерина Л.', xp: 7200 },
  { place: 10, name: 'Алексей Д.', xp: 6800 },
];

const Leaderboard = () => {
  return (
    <div className="leaderboard__page">
        <div className="leaderboard__info">
            <h1 className="leaderboard__title">Лидерборд</h1>
            <p className="leaderboard__description">Топ игроков по опыту</p>
        </div>
        <table className="leaderboard">
            <thead className="leaderboard__head">
                <tr className="leaderboard__row leaderboard__row--header">
                <th className="leaderboard__cell leaderboard__cell--header">Место</th>
                <th className="leaderboard__cell leaderboard__cell--header">Игрок</th>
                <th className="leaderboard__cell leaderboard__cell--header">Опыт(XP)</th>
                </tr>
            </thead>

            <tbody className="leaderboard__body">
                {players.map((player) => (
                <tr key={player.place} className="leaderboard__row">
                    <td className="leaderboard__cell leaderboard__cell--body">{player.place}</td>
                    <td className="leaderboard__cell leaderboard__cell--body">{player.name}</td>
                    <td className="leaderboard__cell leaderboard__cell--body">{player.xp}</td>
                </tr>
                ))}
            </tbody>

            <tfoot className="leaderboard__foot">
                <tr className="leaderboard__row leaderboard__row--footer">
                <td colSpan={3} className="leaderboard__cell leaderboard__cell--footer">
                    <span className="leaderboard__info">Выше место: 5 из 42</span>
                    <button className="leaderboard__button leaderboard__button--share">
                    Поделиться
                    </button>
                </td>
                </tr>
            </tfoot>
        </table>
    </div>
  );
};

export default Leaderboard;