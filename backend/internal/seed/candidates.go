package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }

type seedCandidate struct {
	FullName            string
	Email               string
	Age                 *int
	City                *string
	School              *string
	GraduationYear      *int
	Achievements        *string
	Extracurriculars    *string
	Essay               string
	MotivationStatement *string
}

func SeedCandidates(pool *pgxpool.Pool) error {
	ctx := context.Background()

	var count int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&count)
	if count > 0 {
		log.Printf("Candidates already seeded (%d found), skipping", count)
		return nil
	}

	candidates := []seedCandidate{
		// === STRONG RECOMMEND (5) ===
		{
			FullName: "Aigerim Suleimenova", Email: "aigerim.s@mail.com",
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("NIS PhM Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Founded 'CodeGirls KZ' — a free coding bootcamp for girls in rural Kazakhstan, trained 200+ students across 8 regions. Won 1st place at NURIS National Science Fair 2025. Published research on air quality monitoring using IoT sensors in Eurasian Journal of Applied Sciences."),
			Extracurriculars: ptr("President of School Debate Club (3 years). Volunteer at SOS Children's Villages Almaty. Organizer of TEDxYouth@NIS Almaty 2025."),
			Essay: `When I was 14, I traveled to my grandmother's village in Kyzylorda region for summer break. There, I met Madina, a bright 12-year-old who dreamed of becoming a programmer but had never seen a computer in person. That encounter changed my life.

I came back to Almaty and couldn't stop thinking about Madina. How many others like her existed across Kazakhstan? Kids with enormous potential but zero access. I started CodeGirls KZ from my bedroom — recording coding tutorials in Kazakh, convincing my CS teacher to lend old laptops, and messaging every school principal I could find on social media.

The first session had 4 girls in a dusty classroom in Taldykorgan. By the end of the year, we had expanded to 8 regions. I failed constantly — sponsors rejected me, some schools didn't believe girls needed coding, and one time our entire server crashed during a live session with 50 students. Each failure taught me something. The sponsor rejections taught me to pitch better. The skeptical schools taught me to bring data. The server crash taught me to always have a backup plan.

Now CodeGirls KZ has trained over 200 girls. Three of them won regional olympiads. Madina, the girl from Kyzylorda, just built her first mobile app. Seeing her face on our video call when it worked — that's why I do this.

I want to join inVision U because I've learned to build things from nothing, but I need to learn how to build things that last. I want to understand how inDrive scaled from Yakutsk to the world, because I want to do the same for education access in Central Asia.`,
			MotivationStatement: ptr("I believe education is the most powerful equalizer, and I want to spend my life proving it. inVision U represents exactly the kind of bold, mission-driven thinking that I want to be part of."),
		},
		{
			FullName: "Daulet Kenzhebaev", Email: "daulet.k@mail.com",
			Age: ptrInt(18), City: ptr("Astana"), School: ptr("BIL Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Created 'EcoTrack' — a waste sorting app used by 3 apartment complexes in Astana (500+ users). National Math Olympiad bronze medalist 2024. Led a team of 6 to build a school library management system used by 12 schools."),
			Extracurriculars: ptr("Captain of school basketball team. Mentor at Astana Hub startup incubator for teens. Volunteer translator for UNHCR Kazakhstan."),
			Essay: `My first startup failed spectacularly. At 15, I built a homework-sharing app that I thought would revolutionize education. Instead, teachers reported it as a cheating tool and my principal called my parents. I was mortified.

But that failure was the best thing that happened to me. It taught me that technology without understanding the problem is just noise. I spent the next month talking to teachers — actually listening to them. They didn't need students to share homework. They needed help tracking which students were falling behind.

I rebuilt the app as a teacher's dashboard. It was ugly and buggy, but three teachers started using it. Their feedback shaped every update. This experience taught me the most important lesson of my life: build WITH people, not FOR them.

This principle guided EcoTrack. Before writing a single line of code, I spent two weeks standing next to garbage bins in my apartment complex, watching how people sorted waste. Most didn't sort at all — not because they didn't care, but because the system was confusing. EcoTrack simplified it with visual guides and gamification. Now three complexes use it.

I want to join inVision U because I'm obsessed with solving real problems, and I know I'm just scratching the surface. I need mentors who've built things at scale and peers who challenge me to think bigger.`,
			MotivationStatement: ptr("I want to build technology that serves communities, not just consumers. inVision U's focus on creating leaders who drive change aligns perfectly with my path."),
		},
		{
			FullName: "Madina Orazova", Email: "madina.o@mail.com",
			Age: ptrInt(17), City: ptr("Shymkent"), School: ptr("NIS ChB Shymkent"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Organized city-wide blood donation drive that collected 150+ units. Regional debate champion 2024-2025. Created a mentorship program connecting university students with high schoolers in Turkestan region — 80 pairs matched."),
			Extracurriculars: ptr("Founder and president of Youth Civic Engagement Club. Volunteer at local women's shelter. Writer for school newspaper (published 30+ articles)."),
			Essay: `I grew up watching my mother navigate a system that wasn't built for her. As a single parent in Shymkent, she worked two jobs, dealt with bureaucracy that seemed designed to confuse, and still managed to raise three children who believed they could change the world. She is my definition of leadership — not the loud, charismatic kind, but the quiet, relentless kind that moves mountains one pebble at a time.

Her example taught me that the most important problems aren't always the most visible ones. When I started the Youth Civic Engagement Club, my classmates wanted to organize flashy events. I pushed for something different — understanding our community first. We spent a month interviewing 100 residents of our neighborhood about their biggest daily challenges.

The answer surprised us: loneliness. Not poverty, not infrastructure — loneliness. Elderly people living alone, young mothers isolated at home, teenagers feeling disconnected. We started a simple visiting program. Every Saturday, 15 of us split into groups and visited elderly residents, helped young mothers with groceries, organized game nights for teens.

It wasn't glamorous. No press coverage, no awards. But when Apa Gulmira — an 82-year-old woman who hadn't had a visitor in months — cried and said "I thought everyone forgot about me," I knew we were doing something that mattered.

This is what I want to study at inVision U — not just how to build products, but how to see the invisible problems that technology alone cannot solve.`,
			MotivationStatement: ptr("I want to learn how to combine empathy with innovation. inVision U's mission of creating changemakers resonates deeply with my belief that real impact starts with understanding people."),
		},
		{
			FullName: "Arman Tulegenov", Email: "arman.t@mail.com",
			Age: ptrInt(18), City: ptr("Almaty"), School: ptr("KBTU Lyceum"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Built and sold a Telegram bot for restaurant reservations (2000+ users, sold for $500). National Physics Olympiad silver 2025. Completed Google's Android Development certification at age 16."),
			Extracurriculars: ptr("Open source contributor (3 merged PRs on popular Kotlin libraries). Organizer of AlmatyJS meetup for young developers. Tutor at local orphanage teaching basic computer skills."),
			Essay: `I sold my first piece of software when I was 16. A Telegram bot that let people reserve tables at restaurants in Almaty. It wasn't revolutionary — it was a simple CRUD app with a nice interface. But it worked, and a restaurant owner paid me $500 for it. That money meant less to me than the lesson: someone valued something I built enough to pay for it.

The journey to that point was anything but smooth. I started coding at 13 by copying tutorials from YouTube. My first "app" was a calculator that crashed whenever you divided by zero. I spent six months trying to build a social network (every teenage developer's rite of passage) before realizing I had no idea what I was doing.

What changed everything was contributing to open source. At 15, I found a bug in a Kotlin library I was using, fixed it, and submitted a pull request. When the maintainer — a developer from Berlin — merged my code and wrote "nice fix!", I felt like I had joined a global community. Since then, I've had three PRs merged across different projects.

But my most meaningful work is at the orphanage. Every Sunday, I teach 10 kids basic computer skills. These kids have never had someone sit with them and patiently explain how the internet works. Watching a 10-year-old's eyes light up when they create their first Google Doc — that's more rewarding than any app sale.

I want inVision U to help me bridge these two worlds — the technical excellence of open source and the human impact of education.`,
			MotivationStatement: ptr("I believe the best technology is built by people who understand both code and community. inVision U can help me become that kind of builder."),
		},
		{
			FullName: "Zhanna Bektursynova", Email: "zhanna.b@mail.com",
			Age: ptrInt(17), City: ptr("Aktobe"), School: ptr("Daryn School Aktobe"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Won national essay competition on climate change (1st place among 3000 entries). Founded school recycling program that diverted 2 tons of waste in first year. Selected for FLEX exchange program finalist (top 3%)."),
			Extracurriculars: ptr("Editor-in-chief of school literary magazine. Regional volleyball champion. Volunteer English teacher at community center."),
			Essay: `Aktobe is not where people expect changemakers to come from. It's a small industrial city in western Kazakhstan, known for its chrome factories, not its innovators. Growing up here, I constantly heard "you need to move to Almaty or Astana to do anything meaningful."

I decided to prove that wrong.

When I started our school's recycling program, people laughed. "Recycling? In Aktobe? There's no infrastructure for that." They were right — there wasn't. So I built it. I called every waste management company in the region until I found one willing to partner with us. I convinced our school administration to install sorting bins by presenting data on how much waste our school produced (800 kg per month — I weighed it myself for three months).

The program diverted 2 tons of waste in its first year. More importantly, it changed how 500 students think about consumption. When I overheard a first-grader telling her mother "we need to sort our trash, it's important" — that was my proudest moment.

My essay on climate change that won the national competition wasn't about grand policy proposals. It was about Aktobe — about how the chrome factories affected our air quality, how my grandmother's garden produces less each year, how the Ilek River is shrinking. I wrote about what I see every day, and apparently that specificity resonated with the judges.

I want to attend inVision U because I've proven that impact doesn't require a big city. But I know I need better tools, broader perspective, and a community of people who think like me to scale what I've started.`,
			MotivationStatement: ptr("I want to show that innovation can come from anywhere in Kazakhstan, not just the big cities. inVision U's scholarship model — investing in potential, not privilege — is exactly what I believe in."),
		},
		// === RECOMMEND (10) ===
		{
			FullName: "Nurlan Akhmetov", Email: "nurlan.a@mail.com",
			Age: ptrInt(17), City: ptr("Karaganda"), School: ptr("NIS Karaganda"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Regional math olympiad winner 2024. Created a study group platform used by 50 students. Volunteer at local animal shelter."),
			Extracurriculars: ptr("Chess club captain. Photography club. Participated in Model UN."),
			Essay: `Mathematics has always been my language. While other kids played outside, I solved puzzles. This might sound like a cliché, but for me it was literally true — my father is a math teacher and our dinner table conversations were about number theory.

Winning the regional olympiad was meaningful but expected. What surprised me was what happened after. Classmates who had never spoken to me started asking for help with homework. I realized I could either help them one at a time or build something scalable. So I created a simple platform where students could post questions and others could answer. 50 students use it now — it's nothing fancy, just a Telegram group with organized threads, but it works.

I also volunteer at an animal shelter every weekend. It has nothing to do with math or technology, but it teaches me patience and empathy — two skills I know I need more of. Animals don't care about your GPA.

I want to join inVision U because I'm good at solving problems on paper, but I want to learn how to solve problems in the real world.`,
			MotivationStatement: ptr("I want to combine my analytical skills with real-world impact. inVision U can teach me how to translate mathematical thinking into practical solutions."),
		},
		{
			FullName: "Kamila Sagynbaeva", Email: "kamila.s@mail.com",
			Age: ptrInt(18), City: ptr("Almaty"), School: ptr("Haileybury Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Organized a charity concert that raised 500,000 KZT for children's hospital. School student council vice-president. Completed Stanford's online Machine Learning course."),
			Extracurriculars: ptr("Piano (Grade 8 ABRSM). Debate team. Volunteer at Red Crescent."),
			Essay: `The charity concert I organized last year was the hardest thing I've ever done. Not because of the logistics — though booking a venue, finding performers, and selling tickets was challenging — but because I had to convince people to care.

Most of my classmates thought charity was "something adults do." I had to show them that we could make a real difference. I researched the children's hospital's needs, created presentations with real stories, and personally asked every student in our school to buy a ticket or volunteer.

We raised 500,000 KZT. The hospital used it to buy new equipment for the pediatric ward. When I visited and saw the new monitors, I felt something I'd never felt before — the tangible result of organizing people around a cause.

I've also been exploring machine learning through Stanford's online course. I find the intersection of technology and social good fascinating. Could we use ML to predict which patients need the most urgent care? Could we optimize resource allocation in hospitals? These are questions I want to explore at inVision U.`,
			MotivationStatement: ptr("I'm drawn to inVision U because it values action over credentials. I've learned more from organizing a concert than from any textbook."),
		},
		{
			FullName: "Bekzat Nurmagambetov", Email: "bekzat.n@mail.com",
			Age: ptrInt(17), City: ptr("Astana"), School: ptr("NIS Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Built a weather monitoring station for school using Arduino. Regional informatics olympiad 3rd place. Completed freeCodeCamp full-stack certification."),
			Extracurriculars: ptr("Robotics club member. School newspaper tech columnist. Volunteer at Astana Marathon."),
			Essay: `I built my first Arduino project when I was 14 — a weather station that measured temperature, humidity, and air pressure. It was held together with tape and wires, and the readings were sometimes wildly inaccurate. But it was mine, and it worked (mostly).

That project taught me something important: you don't need permission to build things. I didn't wait for a class assignment or a teacher's instruction. I watched YouTube tutorials, ordered parts with my allowance money, and figured it out through trial and error. When the temperature sensor kept reading 50°C in December, I spent three days debugging before realizing I had a resistor in the wrong place.

Since then, I've been on a continuous learning journey. freeCodeCamp taught me web development, the robotics club taught me teamwork, and the informatics olympiad taught me to think under pressure. Each experience builds on the last.

What I lack is direction. I know I can build things, but I'm not sure what to build. I see problems everywhere — inefficient school systems, environmental waste, bureaucratic processes — but I don't know which ones are worth solving first. I hope inVision U can help me figure that out.`,
			MotivationStatement: ptr("I'm a builder looking for a purpose. inVision U's focus on creating changemakers can help me channel my technical skills toward meaningful problems."),
		},
		{
			FullName: "Dinara Ospanova", Email: "dinara.o@mail.com",
			Age: ptrInt(17), City: ptr("Pavlodar"), School: ptr("School #17 Pavlodar"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Created an Instagram page teaching Kazakh history with 5,000 followers. Won regional creative writing competition. Organized school career day with 10 local professionals."),
			Extracurriculars: ptr("School drama club lead actress. Art club. Volunteer reader at children's library."),
			Essay: `I never thought Instagram could be educational until I created @qazaq_tarihy. It started as a school project — we had to present on Kazakh history, and I thought, why not make it visual? I created infographics about the Kazakh Khanate, short videos explaining nomadic culture, and posts about historical figures nobody talks about.

Within three months, the page had 5,000 followers. People from all over Kazakhstan — students, teachers, even university professors — messaged me saying they learned things they never knew about their own history. A teacher in Atyrau told me she uses my infographics in her classes.

This experience taught me something about communication: people don't engage with information — they engage with stories. When I posted a dry timeline of events, it got 50 likes. When I posted about a 16-year-old warrior queen who defended her tribe against an invasion, it got 2,000 likes. Same history, different storytelling.

I brought this principle to our school career day. Instead of having professionals give boring lectures, I asked each one to share their biggest failure and what they learned from it. The students were captivated. One classmate told me it was "the first school event that felt real."

I want to learn at inVision U how to combine storytelling with technology to make education more engaging and accessible.`,
			MotivationStatement: ptr("I believe that how you tell a story matters as much as the story itself. inVision U can help me learn how to scale impactful storytelling with technology."),
		},
		{
			FullName: "Tamerlan Yessimov", Email: "tamerlan.y@mail.com",
			Age: ptrInt(18), City: ptr("Almaty"), School: ptr("Republican Physics-Math School"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("National programming olympiad participant (top 20). Built a chess engine in Python. Research assistant at IITU machine learning lab."),
			Extracurriculars: ptr("Competitive programming club. Online tutor for math. Runner (completed Almaty half-marathon)."),
			Essay: `Building a chess engine changed how I think about problem-solving. At first, I approached it like a brute-force optimization — evaluate every possible move. My engine was correct but painfully slow. A game against it lasted 40 minutes per move.

The breakthrough came when I studied how grandmasters actually think. They don't evaluate every possibility — they use pattern recognition and heuristics. I implemented alpha-beta pruning and a basic evaluation function based on piece positioning. The engine went from 40 minutes per move to 3 seconds.

This experience parallels my journey in competitive programming. Early on, I tried to solve problems by throwing code at them. Now I spend 70% of my time understanding the problem and 30% coding. The result: I went from not qualifying for nationals to finishing in the top 20.

At the IITU lab, I'm learning that real-world problems are even messier than competitive programming problems. The data is noisy, the requirements change, and "correct" is often subjective. I find this messiness exciting — it means there's always a better solution to find.

I want to attend inVision U to bridge the gap between theoretical problem-solving and building things that help real people.`,
			MotivationStatement: ptr("I'm strong technically but I need to develop my ability to identify problems worth solving. inVision U's entrepreneurial focus is what I need."),
		},
		{
			FullName: "Aisha Muratova", Email: "aisha.m@mail.com",
			Age: ptrInt(17), City: ptr("Shymkent"), School: ptr("Nazarbayev Intellectual School Shymkent"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Founded a peer tutoring network connecting 30 tutors with 100 students. Regional English olympiad gold. Selected for UNESCO Youth Forum Kazakhstan delegate."),
			Extracurriculars: ptr("Student council president. Debate team captain. Volunteer at SOS Children's Village."),
			Essay: `Being student council president taught me that leadership is mostly about listening. When I ran for the position, I had grand plans — new clubs, better facilities, student voice in policy. Reality was different. Most of my time was spent mediating conflicts, organizing mundane events, and sitting in meetings where nothing got decided.

But within this frustration, I found my most impactful project. Students kept complaining about struggling academically but being too embarrassed to ask teachers for help. So I created a peer tutoring network — upperclassmen teaching underclassmen. The key insight was making it social, not academic. We didn't call it "tutoring" — we called it "study hangouts." We met in the cafeteria, not the library. We played music in the background.

30 tutors now help 100 students regularly. Test scores in the participating group improved by an average of 15%. But more importantly, it built connections across grade levels that didn't exist before.

As a UNESCO Youth Forum delegate, I learned that solutions that work in one context often fail in another. What works in Shymkent might not work in Aktau. This taught me the importance of understanding local context before proposing solutions.`,
			MotivationStatement: ptr("I've learned that real leadership is about creating systems, not being in charge. inVision U can help me think systemically about the problems I care about."),
		},
		{
			FullName: "Yerbol Satybaldiev", Email: "yerbol.s@mail.com",
			Age: ptrInt(18), City: ptr("Atyrau"), School: ptr("School #24 Atyrau"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Created a YouTube channel teaching physics in Kazakh (8,000 subscribers). Built simple educational games for younger students. Regional physics olympiad participant."),
			Extracurriculars: ptr("School physics lab assistant. Basketball team. Volunteer at elderly care home."),
			Essay: `In Atyrau, most educational content is in Russian or English. Kazakh-speaking students — especially in smaller towns nearby — struggle to find quality learning materials in their language. I started my YouTube channel to fix this.

My first video was terrible. Bad audio, messy whiteboard, and I stumbled over my words. But I posted it anyway. The comments were encouraging — people said they finally understood topics that had confused them in class. That feedback kept me going.

Now I have 8,000 subscribers and 200+ videos covering the entire high school physics curriculum in Kazakh. I receive messages from students across Kazakhstan — from Turkestan to Semey — thanking me for helping them pass their exams. A girl from a small town in Mangystau region told me my videos were the reason she scored 90 on her UNT physics section.

What I've learned is that access to education isn't just about having schools — it's about having materials in a language you think in. When you learn physics in a language that isn't your own, you're solving two problems simultaneously — the physics problem and the translation problem.

I want inVision U to help me turn this YouTube channel into something bigger — maybe an ed-tech platform for Kazakh-speaking students across Central Asia.`,
			MotivationStatement: ptr("Education in your mother tongue is a right, not a privilege. I want inVision U to help me scale my work beyond YouTube."),
		},
		{
			FullName: "Saltanat Rakhimova", Email: "saltanat.r@mail.com",
			Age: ptrInt(17), City: ptr("Kostanay"), School: ptr("NIS Kostanay"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Created a school mental health awareness campaign reaching 800 students. Regional biology olympiad silver. Published article on teen stress in local newspaper."),
			Extracurriculars: ptr("Peer counselor. Book club organizer. Volunteer at Red Cross."),
			Essay: `Nobody talks about mental health in Kazakh schools. When I tried to start a conversation about it, my teacher told me "Kazakh people are strong, we don't have those problems." That response made me even more determined.

I started small — anonymous surveys among my classmates. The results were alarming: 60% reported significant stress, 40% said they had no one to talk to, and 15% had considered self-harm. I presented these numbers to our school administration. They listened.

Together, we organized a mental health awareness week. We invited a psychologist, set up an anonymous question box, and created a peer support group. The peer counseling program that emerged from this now serves 800 students across three schools.

The hardest part wasn't organizing events — it was fighting stigma. In our culture, admitting you're struggling is seen as weakness. I learned that you can't change minds with data alone. You need stories. When a popular senior athlete shared publicly that he went through depression, it gave permission for others to speak up.

I learned that systemic change requires both data and empathy, both evidence and storytelling. This is what I want to develop further at inVision U.`,
			MotivationStatement: ptr("I want to build systems that support mental health in communities where the concept itself is stigmatized. inVision U's focus on real-world impact draws me in."),
		},
		{
			FullName: "Rustem Torekulov", Email: "rustem.t@mail.com",
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("Lyceum 134 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Led school science fair that attracted 100+ projects. Internship at local IT company (QA testing). Completed CS50 online course."),
			Extracurriculars: ptr("Science club president. Football team goalkeeper. Volunteer math tutor."),
			Essay: `My internship at a small IT company in Almaty lasted only two months, but it changed my perspective completely. I was doing QA testing — clicking through apps, finding bugs, writing reports. It wasn't glamorous, but I saw how software actually gets built: messy, iterative, full of compromises.

The most valuable thing I learned wasn't technical — it was how a team works together under pressure. I watched developers argue about architecture, product managers change requirements mid-sprint, and designers push back on technical limitations. It was chaotic but productive.

This experience made me realize that building something meaningful requires more than coding skills. You need to communicate, negotiate, and sometimes accept that your elegant solution isn't what the customer actually needs.

I brought this mindset back to school when organizing our science fair. Instead of judging projects only on scientific merit, I added categories for "most creative" and "best presented." This encouraged students who weren't traditional "science kids" to participate. We went from 30 projects to over 100.

I'm applying to inVision U because I want to learn how to build products and lead teams — not just write code.`,
			MotivationStatement: ptr("I've seen how real teams build software, and I want to learn how to lead them. inVision U's project-based approach is exactly what I need."),
		},
		// === BORDERLINE (8) ===
		{
			FullName: "Daniyar Aubakirov", Email: "daniyar.a@mail.com",
			Age: ptrInt(17), City: ptr("Astana"), School: ptr("School #55 Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Participated in hackathon (didn't win). Good grades in mathematics. Member of school robotics club."),
			Extracurriculars: ptr("Robotics club. Gaming club. Occasional volunteer."),
			Essay: `I want to study at inVision U because I believe education is the key to success. I have always been interested in technology and want to learn more about programming and entrepreneurship. My school has a robotics club where I learned basic Arduino programming. I also participated in a hackathon where our team tried to build a food delivery app.

I think inVision U is a great opportunity for me because of the scholarship and the quality of education. Kazakhstan needs more tech specialists and I want to be one of them. I am hardworking and motivated, and I believe I can succeed if given the chance.

I don't have many big achievements yet, but I am eager to learn and grow. I think university is where I will find my direction and start building meaningful projects.`,
			MotivationStatement: ptr("I want to become a successful entrepreneur and create technology that helps people. inVision U can give me the skills and network I need."),
		},
		{
			FullName: "Gulnaz Temirgaliyeva", Email: "gulnaz.t@mail.com",
			Age: ptrInt(18), City: ptr("Semey"), School: ptr("School #1 Semey"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Regional English speech competition 2nd place. Helped organize school sports day. Good academic standing."),
			Extracurriculars: ptr("English club. School choir. Helped at mother's small business."),
			Essay: `Growing up in Semey has given me a unique perspective on challenges. Our city has a complicated history, and people here are resilient. I've watched my mother run a small tailoring business for 15 years, working 12-hour days to provide for our family.

Helping her with the business taught me about customer service, time management, and persistence. Sometimes she has more orders than she can handle, and I've thought about how technology could help her manage her workflow better — maybe a simple scheduling app.

I'm good at English and enjoy communicating with people from different backgrounds. At the speech competition, I talked about women entrepreneurs in Kazakhstan, drawing from my mother's experience. The judges said my speech was authentic, which meant a lot to me.

I want to attend inVision U because I see it as a bridge to opportunities that don't exist in Semey. I want to learn skills that I can bring back to my community and use to help people like my mother.`,
			MotivationStatement: ptr("I want to use education to create opportunities for people in smaller cities like Semey, where resources are limited but determination is not."),
		},
		{
			FullName: "Miras Serikbayev", Email: "miras.s@mail.com",
			Age: ptrInt(17), City: ptr("Aktau"), School: ptr("School #14 Aktau"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("School basketball team MVP. Basic Python skills (self-taught). Helped organize beach cleanup."),
			Extracurriculars: ptr("Basketball. Swimming. Some coding projects."),
			Essay: `I'm from Aktau, a city on the Caspian Sea. Life here is different from Almaty or Astana — fewer opportunities, fewer resources, but incredible natural beauty. I love this city, but I know I need to leave it to grow.

I started learning Python last year through online courses. It's hard without a mentor or community, but I'm making progress. I've built a few small programs — a calculator, a to-do list, a simple game. Nothing impressive, but each one teaches me something new.

Basketball is my other passion. Being team MVP taught me about discipline, teamwork, and handling pressure. When the game is on the line and everyone's looking at you to make the shot — that's a feeling that transfers to everything else in life.

I organized a beach cleanup because the Caspian coastline near our school was covered in trash. 30 students showed up and we collected 50 bags of garbage. It felt good to do something tangible for our city.

I want to come to inVision U because I know I have potential but I need the right environment to develop it.`,
			MotivationStatement: ptr("I need a community of driven people to push me to reach my potential. Aktau doesn't have that, but inVision U does."),
		},
		{
			FullName: "Aliya Nurlankyzy", Email: "aliya.n@mail.com",
			Age: ptrInt(17), City: ptr("Taldykorgan"), School: ptr("School #5 Taldykorgan"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Regional art competition 3rd place. Created handmade jewelry sold at local market. Decent academic performance."),
			Extracurriculars: ptr("Art club. Handcraft workshop. Occasional community events."),
			Essay: `Art is how I express myself. I've been drawing since I was 5 and making jewelry since I was 12. My pieces are inspired by Kazakh ornamental patterns — I try to blend traditional designs with modern aesthetics. I sell them at our local market on weekends.

Making and selling jewelry taught me about business in a practical way. I learned to calculate costs, price items, interact with customers, and handle rejection when nobody buys anything. Some weekends are good, others are not. You learn to not take it personally.

I want to study at inVision U because I'm interested in the intersection of design and technology. I've seen how platforms like Etsy allow artisans to reach global markets, and I wonder if something similar could work for Kazakh craftspeople.

I'll be honest — I'm not the strongest student academically. My strength is creativity and persistence. I hope inVision U values those qualities too.`,
			MotivationStatement: ptr("I believe creativity and business sense can create real change, especially for artisans in small cities like mine."),
		},
		{
			FullName: "Olzhas Kurmangaliyev", Email: "olzhas.k@mail.com",
			Age: ptrInt(18), City: ptr("Karaganda"), School: ptr("School #85 Karaganda"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Completed online web development course. School event photographer. Participated in local startup weekend."),
			Extracurriculars: ptr("Photography. Web design hobby projects. Gaming."),
			Essay: `I taught myself web development over the past year using free online resources. I can build basic websites with HTML, CSS, and JavaScript. My proudest project is a portfolio website for my photography — it's simple but clean, and I designed every pixel myself.

Photography and web design share something in common: composition. In both, you're arranging elements to create something visually pleasing and functional. This connection between disciplines fascinates me.

At the local startup weekend in Karaganda, our team pitched a platform for connecting freelance photographers with clients. We didn't win, but the experience of pitching to real entrepreneurs was invaluable. They gave us honest feedback — our idea was too broad, our market research was shallow, and our business model didn't make sense. It was humbling but educational.

I want to apply to inVision U because I want to learn how to turn my skills into something bigger. I can build websites and take photos, but I don't know how to build a business or create something truly impactful.`,
			MotivationStatement: ptr("I have skills but lack direction and mentorship. inVision U can provide both."),
		},
		{
			FullName: "Zarina Yessenova", Email: "zarina.y@mail.com",
			Age: ptrInt(17), City: ptr("Shymkent"), School: ptr("School #35 Shymkent"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("School play lead role. Good grades. Helped with family business accounting."),
			Extracurriculars: ptr("Drama club. Dance. Helping family."),
			Essay: `I've been helping my father with his small grocery store since I was 14. I do the accounting — tracking expenses, counting inventory, and calculating profits. It's not glamorous, but it gave me practical business knowledge that most of my classmates don't have.

Our store serves a neighborhood of about 200 families. I know most of our customers by name. This personal connection is something big supermarkets can never replicate. But it's also a weakness — our business depends on my father being there every day, 7 days a week.

I've thought about how technology could help small stores like ours compete with chains. A simple inventory management system, a loyalty program, online ordering — these things exist for big businesses but are too expensive for small ones.

Drama club is my escape from the routine of school and the store. Playing different characters teaches me empathy — to see the world through someone else's eyes. My favorite role was playing a teacher who realizes she's been too harsh with her students. The irony wasn't lost on me.`,
			MotivationStatement: ptr("I want to help small businesses in Kazakhstan compete in the digital age. My family's store taught me the problems; I need inVision U to teach me the solutions."),
		},
		{
			FullName: "Erlan Kazhimov", Email: "erlan.k@mail.com",
			Age: ptrInt(18), City: ptr("Almaty"), School: ptr("School #125 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("School sports day organizer. Basic coding knowledge. Average academic performance."),
			Extracurriculars: ptr("Football. Some coding. Student council member."),
			Essay: `I organized our school's sports day last year. It was a lot of work — scheduling events, coordinating teams, setting up equipment, and dealing with complaints when things didn't go as planned. But it was worth it when I saw 300 students having fun and competing in a friendly atmosphere.

I'm interested in technology and have been learning to code on my own, though I'm still at the beginner level. I can write basic Python programs and I've started learning about databases. It's challenging without a teacher, but online resources help.

I think inVision U is an amazing opportunity because of the scholarship and the chance to learn from experienced mentors. I come from a middle-class family and couldn't afford private university. This scholarship would change my life.

I'm honest about my limitations — I'm not the smartest student or the most talented coder. But I'm reliable, I work hard, and I care about doing things right.`,
			MotivationStatement: ptr("inVision U represents a life-changing opportunity for me. I promise to make the most of it and give back to my community."),
		},
		{
			FullName: "Inkar Baimukhanova", Email: "inkar.b@mail.com",
			Age: ptrInt(17), City: ptr("Oral"), School: ptr("School #3 Oral"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Won school English essay competition. Tutored younger students in math. Participated in city volunteering events."),
			Extracurriculars: ptr("English club. Math tutoring. City volunteering."),
			Essay: `I come from Oral, a city in western Kazakhstan that most people haven't heard of. We're close to the Russian border and far from the centers of opportunity in Kazakhstan. But distance doesn't diminish ambition.

I tutor younger students in math because I believe everyone deserves help when they're struggling. I remember how lost I felt in 7th grade when algebra suddenly got hard. A classmate helped me, and that kindness stuck with me. Now I pass it on.

Winning the English essay competition was unexpected. I wrote about the importance of learning languages in a globalized world — how speaking English opened doors for me through online courses and international pen pals. The judges liked my practical perspective.

I want to attend inVision U because I've done everything I can with the resources available to me in Oral. To grow further, I need new challenges, new people, and new perspectives.`,
			MotivationStatement: ptr("Distance from major cities shouldn't limit potential. I want inVision U to help me prove that talent exists everywhere."),
		},
		// === NOT RECOMMENDED (7) ===
		{
			FullName: "Askar Jumagulov", Email: "askar.j@mail.com",
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("School #78 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("None significant"),
			Extracurriculars: ptr("Gaming"),
			Essay: `I want to go to inVision U because it offers free education and I think it would be a good opportunity. I heard about it from a friend and decided to apply. I like technology and want to learn programming. I think I would be a good student because I am smart and learn quickly. Please give me a chance.`,
			MotivationStatement: ptr("I want free education and a good career."),
		},
		// AI-generated essay (HIGH risk)
		{
			FullName: "Sanzhar Omarov", Email: "sanzhar.o@mail.com",
			Age: ptrInt(18), City: ptr("Astana"), School: ptr("School #100 Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Member of school science club. Participated in city quiz tournament."),
			Extracurriculars: ptr("Science club. Quiz club. Reading."),
			Essay: `In today's rapidly evolving technological landscape, the intersection of artificial intelligence and education presents unprecedented opportunities for transformative change. As a passionate advocate for leveraging cutting-edge innovations to address systemic challenges, I firmly believe that inVision U represents the pinnacle of forward-thinking educational paradigms.

Throughout my academic journey, I have consistently demonstrated an unwavering commitment to excellence and a profound dedication to making a meaningful impact in my community. My experiences have equipped me with a diverse skill set that encompasses critical thinking, creative problem-solving, and effective communication.

The holistic approach to education at inVision U resonates deeply with my personal philosophy of lifelong learning and continuous improvement. I am particularly drawn to the institution's emphasis on fostering entrepreneurial mindsets and cultivating leadership qualities that are essential for navigating the complexities of the 21st century.

Furthermore, I am eager to contribute to the vibrant tapestry of diverse perspectives that characterizes the inVision U student body. I believe that my unique background and experiences will enrich the collaborative learning environment and foster cross-cultural understanding.

In conclusion, I am confident that my passion for innovation, combined with my track record of academic achievement and community engagement, makes me an ideal candidate for inVision U. I look forward to the opportunity to grow, learn, and make a lasting impact through this transformative educational experience.`,
			MotivationStatement: ptr("I am driven by an insatiable curiosity and an unwavering commitment to leveraging technology for the betterment of society. inVision U's innovative curriculum and distinguished faculty make it the ideal institution for actualizing my aspirations."),
		},
		// AI-generated essay (HIGH risk)
		{
			FullName: "Merey Iskakov", Email: "merey.i@mail.com",
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("School #45 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Good grades. Participated in school olympiad. Member of debate club."),
			Extracurriculars: ptr("Debate club. Chess. Reading."),
			Essay: `As we stand on the precipice of a new era in human development, the importance of quality education cannot be overstated. inVision U, with its revolutionary approach to nurturing the next generation of leaders and innovators, stands as a beacon of hope in the educational landscape of Central Asia.

My journey has been characterized by a relentless pursuit of knowledge and a deep-seated desire to effect positive change in the world around me. From my earliest years, I have been fascinated by the power of technology to solve complex problems and improve lives. This fascination has driven me to explore various domains, from computer science to social entrepreneurship, always seeking to expand my horizons and deepen my understanding.

The synergistic combination of academic rigor, practical experience, and values-based leadership development that inVision U offers is precisely what I need to catalyze my growth as a future changemaker. I am particularly excited about the prospect of collaborating with like-minded individuals from diverse backgrounds, as I firmly believe that the most innovative solutions emerge from the intersection of different perspectives and disciplines.

I am confident that my intellectual curiosity, combined with my commitment to community service and my entrepreneurial spirit, will allow me to both contribute to and benefit from the exceptional learning community at inVision U.`,
			MotivationStatement: ptr("My aspiration to become a catalyst for positive change in society aligns seamlessly with inVision U's mission to develop leaders who will shape the future."),
		},
		// Medium AI risk
		{
			FullName: "Nurislam Bekzhanov", Email: "nurislam.b@mail.com",
			Age: ptrInt(17), City: ptr("Shymkent"), School: ptr("School #20 Shymkent"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Participated in a few school events. Average grades."),
			Extracurriculars: ptr("Football. Sometimes helps at mosque."),
			Essay: `I am writing to express my sincere interest in the inVision U program. Education has always been important to me, and I believe this opportunity will help me achieve my dreams.

To be honest, I haven't done many outstanding things yet. I play football with my friends, I sometimes help at our local mosque, and I try to get good grades. I know this doesn't sound impressive compared to other applicants.

But I want to change. I feel like I have potential that hasn't been unlocked yet. In my neighborhood, there aren't many role models or opportunities. Most of my friends plan to work in construction or trade after school. There's nothing wrong with that, but I want something different.

I hope inVision U will give me the chance to discover what I'm capable of. I promise to work hard if accepted.`,
			MotivationStatement: ptr("I need a chance to prove myself. My environment doesn't offer many opportunities, but I believe I can do great things with the right support."),
		},
		// Medium AI risk
		{
			FullName: "Ayaulym Serikova", Email: "ayaulym.s@mail.com",
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("School #60 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("School honor roll. Participated in Model UN. Basic website projects."),
			Extracurriculars: ptr("Model UN. School newspaper. Instagram blogging."),
			Essay: `The modern world requires leaders who can navigate complexity with both analytical rigor and creative thinking. I believe I embody this combination. Through my participation in Model UN, I developed strong public speaking and negotiation skills that have shaped my approach to problem-solving.

On a more personal note, I've always been curious about technology. Last summer, I tried to build a website for our school newspaper. It was basic — just HTML and CSS — but I was proud of it. My teacher said it was "good for a first try," which I'll take as a compliment.

I also run a small Instagram blog about student life in Almaty. Nothing big — just 500 followers — but it taught me about content creation and audience engagement.

I think inVision U is the perfect place for someone like me who has many interests but hasn't found their focus yet. I need guidance and mentorship to channel my energy in the right direction.`,
			MotivationStatement: ptr("I'm at a crossroads in my life and inVision U can help me choose the right path. I'm eager to learn and grow."),
		},
		// Medium AI risk
		{
			FullName: "Temirlan Zhumabekov", Email: "temirlan.z@mail.com",
			Age: ptrInt(18), City: ptr("Kyzylorda"), School: ptr("School #7 Kyzylorda"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Helped organize school concert. Basic English knowledge. Interested in business."),
			Extracurriculars: ptr("Music. Helping family with shop. Reading business books."),
			Essay: `My father has a small shop in Kyzylorda bazaar. I've been helping him since I was young. He taught me that business is about relationships — knowing your customers, being honest, and showing up every day even when you don't want to.

In today's digital economy, however, traditional business practices must evolve to remain competitive. The integration of technology into small business operations represents a critical pathway toward sustainable growth and enhanced market positioning. I believe that gaining expertise in digital transformation strategies would enable me to revolutionize my family's business model and contribute to the broader economic development of my region.

I want to learn at inVision U because I see how the world is changing and I don't want my city to be left behind. Kyzylorda needs people who understand both traditional business and new technology.`,
			MotivationStatement: ptr("I want to modernize business in Kyzylorda. My father's shop taught me the basics, but I need formal education to go further."),
		},
		{
			FullName: "Darkhan Moldabekov", Email: "darkhan.m@mail.com",
			Age: ptrInt(18), City: ptr("Astana"), School: ptr("School #30 Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Nothing notable"),
			Extracurriculars: ptr("Video games. Social media."),
			Essay: `I want to go to university for free thats why I am applying to inVision U. My grades are okay but not great. I like computers and games. I think I could learn programming if someone taught me properly. My school doesn't have good CS classes so I havent had the chance to learn much.

I live with my mom in Astana. She wants me to get higher education but we cant afford regular university. This scholarship is our best option.

I dont have achievements to list but I think potential matters more than past accomplishments. Everyone starts somewhere.`,
			MotivationStatement: ptr("Need scholarship for education. Will try hard."),
		},
	}

	for _, c := range candidates {
		_, err := pool.Exec(ctx,
			`INSERT INTO candidates (full_name, email, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'pending')`,
			c.FullName, c.Email, c.Age, c.City, c.School, c.GraduationYear,
			c.Achievements, c.Extracurriculars, c.Essay, c.MotivationStatement,
		)
		if err != nil {
			log.Printf("Failed to seed candidate %s: %v", c.FullName, err)
		}
	}

	log.Printf("Seeded %d candidates", len(candidates))
	return nil
}
