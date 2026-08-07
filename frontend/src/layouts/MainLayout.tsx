import { Outlet } from "react-router-dom"

const MainLayout = () => {
    return (
        <div>
            <nav>
                <div>Ссылки</div>
                <div>Выход</div>
            </nav>
            <main>
                <Outlet />
            </main>
        </div>
    )
}

export default MainLayout;