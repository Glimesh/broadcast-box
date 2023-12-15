import { useContext } from 'react';
import { Link, Outlet } from 'react-router-dom'
import { CinemaModeContext } from '../player';

const Header = () => {
  const { cinemaMode } = useContext(CinemaModeContext);
  const navbarEnabled = !cinemaMode;
  return (
    <div>
      {navbarEnabled && (
        <nav className='bg-gray-800 p-2 mt-0 fixed w-full z-10 top-0'>
          <div className='container mx-auto flex flex-wrap items-center'>
            <div className='flex flex-1 text-white font-extrabold'>
              <Link to="/" className='font-light leading-tight text-2xl'>
                Broadcast Box
              </Link>
            </div>
          </div>
        </nav>
      )}

      <main className={`${navbarEnabled && "pt-20 md:pt-24"}`}>
        <Outlet />
      </main>

      <footer className="mx-auto px-2 container py-6">
        <ul className="flex items-center justify-center mt-3 text-sm:mt-0 space-x-4">
          <li>
            <a href="https://github.com/Glimesh/broadcast-box" className="hover:underline">GitHub</a>
          </li>
          <li>
            <a href="https://pion.ly" className="hover:underline">Pion</a>
          </li>
          <li>
            <a href="https://glimesh.tv" className="hover:underline">Glimesh</a>
          </li>
        </ul>
      </footer>

    </div>
  )
}

export default Header
