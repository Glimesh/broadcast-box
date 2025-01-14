export default function ErrorHeader({ children: error }) {
  return (
    <p className={'bg-red-700 text-white text-lg ' +
      'text-center p-5 rounded-t-lg whitespace-pre-wrap'
    }>
      {error}
    </p>
  )
}


