import { Breadcrumbs, Anchor } from '@mantine/core';
import { Link } from 'react-router-dom';

export function YBBreadCrumbComponent() {
  const p = window.location.pathname.split("/")
  const crumbItems = [
    {
      title: "Home",
      href: "/",
    }
  ]

  if (p.length > 1 && p[1] !== "") {
    crumbItems.push({
      title: decodeURIComponent(window.location.pathname.split("/")[1]),
      href: "",
    })
  }

  const items = crumbItems.map((item,index) =>
    (item.href === "")?item.title:
      <Anchor component={Link} to={item.href} key={index}>
        {item.title}
      </Anchor>
  )

  return (
    <Breadcrumbs mt="1em">{items}</Breadcrumbs>
  )
}