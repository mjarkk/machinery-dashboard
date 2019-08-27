// From some quick research the react start template doesn't support detection for if ran in production so i'll use this "hack" here
// When running the webapp in development mode the port will be 3000
export const get = async route => {
  return f(route)
}

export const f = async (route, options) => {
  const res = await fetch(`/api${route}`, options)
  return await res.json()
}
