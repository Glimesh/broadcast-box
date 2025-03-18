import { useContext } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom'
import { CinemaModeContext } from '../player';

const Header = () => {
  const location = useLocation();
  const streamKey = location.pathname.split('/').pop();
  const isStreamPage = location.pathname !== '/' && !location.pathname.startsWith('/publish');
  
  const { cinemaMode } = useContext(CinemaModeContext);
  const navbarEnabled = !cinemaMode;
  
  // Title case function
  const toTitleCase = (str) => {
    return str.replace(/\w\S*/g, txt => txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase());
  };
  
  return (
    <div>
      {navbarEnabled && (
        <nav className='bg-[var(--color-primary)] px-2 py-2 mt-0 fixed w-full z-10 top-0 shadow-[var(--shadow-md)]'>
          <div className='w-full flex flex-wrap items-center'>
            <div className='flex flex-1'>
              <Link to="/" className='font-bold tracking-tight text-xl flex items-center text-white hover:text-white'>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5 mr-2">
                  <path d="M4.5 4.5a3 3 0 0 0-3 3v9a3 3 0 0 0 3 3h8.25a3 3 0 0 0 3-3v-9a3 3 0 0 0-3-3H4.5ZM19.94 18.75l-2.69-2.69V7.94l2.69-2.69c.944-.945 2.56-.276 2.56 1.06v11.38c0 1.336-1.616 2.005-2.56 1.06Z" />
                </svg>
                {isStreamPage ? toTitleCase(streamKey) : 'Broadcast Box'}
              </Link>
            </div>
          </div>
        </nav>
      )}

      <main className={`${navbarEnabled && "pt-[3.5rem]"}`}>
        <Outlet />
      </main>
    </div>
  )
}

export default Header
