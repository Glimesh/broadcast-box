import { Outlet, useNavigate } from 'react-router-dom'

const Header = () => {
  const navigate = useNavigate()

  return (
    <div>
      <nav className='bg-gray-800 p-2 mt-0 fixed w-full z-10 top-0'>
        <div className='container mx-auto flex flex-wrap items-center'>
          <div className='flex flex-1 text-white font-extrabold'>
            <a href="/" className='font-light leading-tight text-2xl'>
              Broadcast Box
            </a>
          </div>
          <div className='flex content-center justify-between md:w-1/2 md:justify-end'>
            <ul className='list-reset flex justify-between flex-1 md:flex-none items-center'>
              <li className=''>
                <button
                  className='py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75'
                  onClick={() => { navigate('/') }}
                >
                  Go Home
                </button>
              </li>
            </ul>
          </div>
        </div>
      </nav>

      <main className='pt-20 md:pt-24'>
        <div className='mx-auto px-2 container'>
          <Outlet />
        </div>
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
