import { Outlet, useNavigate } from 'react-router-dom'

const Header = () => {
  const navigate = useNavigate()

  return (
    <div>
      <nav className='bg-gray-800 p-2 mt-0 fixed w-full z-10 top-0'>
        <div className='container mx-auto flex flex-wrap items-center'>
          <div className='flex w-full md:w-1/2 justify-center md:justify-start text-white font-extrabold'>
            <span className='text-2xl pl-2'>
              Broadcast Box
            </span>
          </div>
          <div className='flex w-full pt-2 content-center justify-between md:w-1/2 md:justify-end'>
            <ul className='list-reset flex justify-between flex-1 md:flex-none items-center'>
              <li className='mr-3'>
                <button
                  className='bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded'
                  onClick={() => { navigate('/') }}
                >
                  Select Stream
                </button>
              </li>
            </ul>
          </div>
        </div>
      </nav>

      <main className='pt-36'>
        <div className='mx-auto w-6/12'>
          <Outlet />
        </div>
      </main>
      <footer />
    </div>
  )
}

export default Header
