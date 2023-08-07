import { ReactNode } from 'react';
import { Breadcrumb } from 'antd';

interface PathElement {
  title: string|JSX.Element
}

export default function YBBreadCrumb() {
  let p = window.location.pathname.split("/")
  let items: PathElement[] = [
    {
      title: <a href="/">Home</a>,
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