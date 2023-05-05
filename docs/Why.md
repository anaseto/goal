# Why a new array programming language?

Short answer is, the choice of features is new, so is that not a reason enough?
:-)

The long answer is:

* It's easily embeddable and extensible in Go. This means easy access to one of
  the most featured standard libraries in the wild. Go is also easier than C
  and provides us with good garbage collection out of the box, which is handy.
* I've had a strong liking for array programming languages since quite a few
  years, but I never managed to like array-based string handling, which is kind
  of an essential issue for many common scripting purposes.  Unicode and UTF-8
  don't play that well with the array vision. Sure, array-based text-handling
  provides very smart solutions for some specific kinds of tasks (BQN's a good
  example of that), but that's it. Text is complicated, there is no
  straightforward mapping between “character” (a somewhat ill-defined notion
  that is maybe best approached by graphemes) and bytes or code points. For
  example, some abstract characters cannot be encoded by a single code point,
  and some abstract characters have more than one possible encoding. That's why
  I feel a scripting language should consider strings as a whole by default.
* Flexible string quoting with interpolation is nice.
* While I do feel the appeal and beauty of tacit code, I'm not completely sold
  on it because of the need to constantly switch between both tacit and
  explicit styles.  Lambdas with {} notation and default arguments (x,y,z) are
  already very concise. Having only the explicit style frees the mind of the
  cognitive load of having to choose between both.
* I wanted both “ASCII is easy to type” and “no digraphs”. Actually, «, » and ¿
  are three exceptions that prove the ASCII rule, but they each have an ASCII
  keyword counterpart, and I have direct access to them on my bépo keyboard
  layout :-)
* I wanted some of BQN's primitives, but leaving the multi-dimensional
  complexity out, like K.
* Last, but not least, I had some previous experience with some aspects of
  compilation, but I never had written a whole bytecode interpreter from
  scratch before: it's a great and fun experience!
