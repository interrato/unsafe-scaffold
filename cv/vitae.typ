#let arancione = rgb("#974a1f")

#let vitae(
  fullname: none, // required
  date: none, // required
  logo: none,
  doc,
) = {
  assert.ne(
    fullname,
    none,
    message: "fullname: expected string or content, found none",
  )

  set document(
    title: fullname,
    author: fullname,
    description: "Curriculum Vitae",
    keywords: ("curriculum vitae", "cv", "resume"),
    date: date,
  )
  set par(justify: true, justification-limits: (
    tracking: (min: -0.01em, max: 0.02em),
  ))
  set strong(delta: 200)
  set page(paper: "a4", margin: 1.875cm)
  set text(font: "EB Garamond", weight: "regular", size: 12pt)

  show heading: it => box(it)
  show link: set text(fill: arancione)
  show title: set text(weight: "medium")

  if logo == none {
    title()
  } else {
    grid(
      columns: (1fr, 1fr),
      align: (left + bottom, right + bottom),
      title(), logo,
    )
  }

  line(length: 100%, stroke: 0.5pt)

  doc
}

#let section(name, body) = {
  set block(below: 0.6cm)
  show heading: set text(weight: "regular", size: 12pt)
  show heading: it => smallcaps(it)
  grid(
    columns: (3.6cm, 1fr),
    column-gutter: 0.6cm,
    align: (top + right, top + left),
    text(hyphenate: false)[= #name], body,
  )
}

#let mono(body) = {
  set text(font: "Fira Code", size: 9.6pt)
  body
}
