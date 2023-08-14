import { Breadcrumb } from 'antd';
import { Link } from 'react-router-dom'

interface PathElement {
  title: string|JSX.Element
}

export function YBBreadCrumb() {
  let p = window.location.pathname.split("/")
  let items: PathElement[] = [
    {
      title: <Link to="/">Home</Link>,
    }
  ]

  if (p.length > 1 && p[1] !== "") {
    items.push({
      title: window.location.pathname.split("/")[1],
    })
  }
  return (
    <Breadcrumb
        items={items}
      />
  )
}