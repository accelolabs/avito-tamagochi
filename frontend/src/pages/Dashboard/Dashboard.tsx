import './Dashboard.css'
import kita from '../../assets/teen.png';

const Dashboard = () => {
    return (
        <div className="dashboard-page">
            <div className="dashboard">
                <div className="dashboard__pet-card">
                    <div className="dashboard__pet-image">
                         <img src={kita} alt="Кита подросток" />
                    </div>
                    <div className="dashboard__pet-name">K1-T4</div>
                    <div className="dashboard__pet-phase">Подросток</div>
                    <div className="dashboard__progress-bars">
                        <div className="dashboard__progress">
                            <div className="dashboard__progress-info">
                                <span className="dashboard__progress-label">ХР (Опыт)</span>
                                <span className="dashboard__progress-value">0 / 1000</span>
                            </div>
                            <div className="dashboard__progress-bar">
                                <div className="dashboard__progress-fill dashboard__progress-xp" style={{ width: '0%' }}></div>
                            </div>
                        </div>
                        <div className="dashboard__progress">
                            <div className="dashboard__progress-info">
                                <span className="dashboard__progress-label">Батарея</span>
                                <span className="dashboard__progress-value">85 %</span>
                            </div>
                            <div className="dashboard__progress-bar">
                                <div className="dashboard__progress-fill dashboard__progress-battery" style={{ width: '85%' }}></div>
                            </div>
                        </div>
                    </div>
                    <button className="dashboard__charge">
                        <svg width="16" height="20" viewBox="0 0 16 20" xmlns="http://www.w3.org/2000/svg">
                            <path d="M6.55 16.2L11.725 10H7.725L8.45 4.325L3.825 11H7.3L6.55 16.2V16.2M4 20L5 13H0L9 0H11L10 8H16L6 20H4V20M7.775 10.25V10.25V10.25V10.25V10.25V10.25V10.25V10.25" fill="currentColor"/>
                        </svg>
                        Зарядить
                    </button>
                </div>
                <div className="dashboard__stats">
                    <div className="dashboard__stat-card">
                        <h2>ВЫПОЛНЕНО ЗАДАНИЙ</h2>
                        <p>2 / 3</p>
                        <svg className="dashboard__stat-icon dashboard__stat-icon--bottom" width="68" height="68" viewBox="0 0 68 68" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M25.9561 0C24.0783 0 22.5561 1.52223 22.5561 3.4V14.7677L13.0339 9.27007C11.4077 8.3312 9.32827 8.8884 8.3894 10.5146L0.456083 24.2555C-0.482801 25.8817 0.0743747 27.9611 1.70057 28.9L11.2009 34.3851L1.70057 39.8701C0.0743726 40.809 -0.482803 42.8884 0.456081 44.5146L8.3894 58.2555C9.32833 59.8817 11.4077 60.4389 13.0339 59.5L22.5561 54.0024V64.6C22.5561 66.4778 24.0783 68 25.9561 68H41.8227C43.7005 68 45.2227 66.4778 45.2227 64.6V54.0275L54.7013 59.5C56.3275 60.4389 58.4069 59.8817 59.3458 58.2555L67.2793 44.5146C68.218 42.8884 67.6607 40.809 66.0347 39.8701L56.5343 34.3851L66.0347 28.9C67.6607 27.9611 68.218 25.8817 67.2793 24.2555L59.3458 10.5146C58.4069 8.8884 56.3275 8.3312 54.7013 9.27007L45.2227 14.7425V3.4C45.2227 1.52223 43.7005 0 41.8227 0H25.9561Z" fill="#965EEB"/>
                        </svg>
                    </div>
                    <div className="dashboard__stat-card">
                        <h2>ДОСТУПНО НАГРАД</h2>
                        <p>1</p>
                        <svg className="dashboard__stat-icon dashboard__stat-icon--top" width="99" height="81" viewBox="0 0 99 81" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M91.5735 42.9163L85.5929 30.5976C84.1664 27.6584 81.7076 25.3504 78.6859 24.1148L39.2486 7.99983C36.2256 6.76431 32.8521 6.68915 29.7742 7.78883L16.8817 12.3956C16.016 12.7045 15.2275 13.198 14.5706 13.8418C13.9137 14.4857 13.4042 15.2645 13.0773 16.1245C12.7503 16.9845 12.6138 17.9052 12.6771 18.8227C12.7405 19.7403 13.0022 20.6328 13.4441 21.4387L34.5996 59.9967C35.64 61.8924 37.2988 63.3725 39.2992 64.1899C41.2996 65.0073 43.5202 65.1125 45.5905 64.4879L87.691 51.7777C88.571 51.5123 89.3829 51.0587 90.0709 50.4483C90.7589 49.8379 91.3066 49.0855 91.676 48.2428C92.0454 47.4001 92.2275 46.4874 92.2099 45.5677C92.1923 44.6481 91.9751 43.7434 91.5735 42.9163Z" fill="#00AAFF"/>
                        </svg>
                    </div>
                    <div className="dashboard__stat-card">
                        <h2>МЕСТО В ЛИДЕРБОРДЕ</h2>
                        <p>5</p>
                        <svg className="dashboard__stat-icon dashboard__stat-icon--bottom" width="94" height="58" viewBox="0 0 94 58" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M1.69326 12.6278C3.8766 7.54361 7.98534 3.53576 13.116 1.48539C18.2468 -0.564991 23.9795 -0.49001 29.0538 1.69385C44.5099 8.33821 60.6152 13.3508 77.109 16.6507C79.7926 17.1833 82.3447 18.2405 84.6198 19.7617C86.8949 21.2829 88.8491 23.2383 90.37 25.5163C91.891 27.7944 92.949 30.3504 93.4834 33.0385C94.0184 35.7266 94.0198 38.4941 93.4876 41.183C92.9547 43.8719 91.8988 46.4296 90.3799 48.7099C88.8611 50.9903 86.909 52.9487 84.6353 54.4733C82.361 55.9979 79.8102 57.059 77.1274 57.5958C74.4445 58.1326 71.6824 58.1347 68.999 57.602C49.6437 53.7293 30.7447 47.8454 12.6084 40.0455C10.0946 38.9643 7.81819 37.3973 5.90917 35.434C4.00013 33.4709 2.49583 31.15 1.48225 28.604C0.468659 26.0579 -0.0343792 23.3366 0.00182403 20.5954C0.0380381 17.8542 0.612771 15.1468 1.69326 12.6278Z" fill="#02D15C"/>
                        </svg>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;