#import "vitae.typ": vitae
#import "vitae.typ": mono, section

#show: vitae.with(
  fullname: "Simone Ragusa",
  date: datetime(year: 2026, month: 6, day: 29),
  logo: move(dy: 0.35em, image("logo.svg", width: 1.08em)),
)

#section("Contact")[
  Email: #mono(link("mailto:simone@interrato.dev")) \
  Website: #mono(link("https://interrato.dev")) \
  Bluesky: #mono(link("https://bsky.app/profile/interrato.dev")[\@interrato.dev]) \
]

#section("Interests")[
  Applied cryptography, high-assurance cryptography, web security, online
  privacy, software reproducibility.
]

#section("Education")[
  *University of Padua*, Padua, Italy

  #box(inset: (left: 0.3cm))[
    Master's degree in Cybersecurity, September 2025 #h(1fr) 110 cum laude \
    #box(inset: (left: 0.3cm))[
      - #link("https://hdl.handle.net/20.500.12608/91818")[
          "Fuzzy Searchable Symmetric Encryption: Design and Implementation of a
          Novel Scheme Toward Real-World Applications"
        ]
      - Thesis supervisor: Nicola Laurenti
    ]
  ]

  #box(inset: (left: 0.3cm))[
    Bachelor's degree in Computer Engineering, March 2022 #h(1fr) 101 out of 110 \
    #box(inset: (left: 0.3cm))[
      - Courses in Computer Architecture, Software Engineering,
        Telecommunications, and Electronics
    ]
  ]
]

#section("Professional Experience")[
  *TEXA S.p.A.*, #mono(link("https://www.texa.it")) \
  _Cybersecurity Engineer_ #h(1fr) *June 2022 -- August 2023* \
  Drafted operating procedures providing guidelines for secure software
  development. Performed vulnerability assessments and penetration testing on
  several web applications. Developed a background service to automate file
  signing and exchange via SFTP over the ENX network. Designed the new internal
  cryptographic key inventory to promote careful scoping and effective key
  rotation.
]

#section("Teaching")[
  *Information Security*, University of Padua \
  _Teaching Assistant_ #h(1fr) *2026* \
  Designed hands-on laboratory projects including the implementation of
  symmetric encryption primitives and padding-oracle attacks, hash functions,
  message authentication codes and timing side-channel attacks, digital
  signatures, and key-exchange protocols. Graded student submissions.
]

#section("Other Activities")[
  *Research*

  #box(inset: (left: 0.3cm))[
    _Searchable Symmetric Encryption_ #h(1fr) *November 2025 -- Present* \
    #box(inset: (left: 0.3cm))[
      - Research grant by University of Padua
    ]
  ]
]

#section("Certifications")[
  *Cryptography I*, Coursera \
  #mono(link("https://coursera.org/verify/P8LLQSUHS6GJ")) #h(1fr) *April 2021*
]

#section("Programming Languages")[
  *Strong:* Go, Python, Bash \
  *Intermediate:* C, C++, Nix, JavaScript \
  *Basic:* Lua, PHP, Java, Kotlin, Rust
]

#section("Tools")[
  *Strong:* (Neo)Vim, Typst \
  *Intermediate:* Git, Linux, LaTeX \
  *Basic:* Burp Suite
]
