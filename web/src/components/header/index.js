import { Outlet, useNavigate } from 'react-router-dom'

const Header = () => {
  const navigate = useNavigate()

  return (
    <div>
      <nav className='bg-gray-800 p-2 mt-0 fixed w-full z-10 top-0'>
        <div className='container mx-auto flex flex-wrap items-center'>
          <div className='flex w-full md:w-1/2 justify-center md:justify-start text-white font-extrabold'>
            <a href="/" className='font-light leading-tight text-2xl'>
              Broadcast Box
            </a>
          </div>
          <div className='flex w-full content-center justify-between md:w-1/2 md:justify-end'>
            <ul className='list-reset flex justify-between flex-1 md:flex-none items-center'>
              <li className=''>
                <button
                  className='py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75'
                  onClick={() => { navigate('/') }}
                >
                  Select Stream
                </button>
              </li>
            </ul>
          </div>
        </div>
      </nav>

      <main className='pt-24'>
        <div className='mx-auto container'>
          <Outlet />
        </div>
      </main>

      <footer className="mx-auto container md:flex md:items-center md:justify-between py-6">
        <span className="text-sm sm:text-center">Provided with love, or something... Do we have a slogan?</span>
        <ul className="flex flex-wrap items-center mt-3 text-sm:mt-0 space-x-4  ">
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
