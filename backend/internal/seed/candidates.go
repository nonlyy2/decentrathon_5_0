package seed

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }

type seedCandidate struct {
	FullName            string
	Email               string
	Phone               *string
	Telegram            *string
	Age                 *int
	City                *string
	School              *string
	GraduationYear      *int
	Achievements        *string
	Extracurriculars    *string
	Essay               string
	MotivationStatement *string
	Major               *string
}

func SeedCandidates(pool *pgxpool.Pool, force bool) error {
	ctx := context.Background()

	var count int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&count)
	if count > 0 && !force {
		log.Printf("Candidates already seeded (%d found), skipping. Use --force-seed to override.", count)
		return nil
	}
	if force && count > 0 {
		log.Printf("Force-seeding: truncating %d existing candidates...", count)
		if _, err := pool.Exec(ctx, `TRUNCATE candidates CASCADE`); err != nil {
			return fmt.Errorf("failed to truncate candidates: %w", err)
		}
	}

	candidates := []seedCandidate{
		// === STRONG RECOMMEND (5) ===
		{
			FullName: "Aigerim Suleimenova", Email: "aigerim.s@mail.com", Phone: ptr("+7 707 123 4567"), Telegram: ptr("@aigerim_s"),
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
			FullName: "Daulet Kenzhebaev", Email: "daulet.k@mail.com", Phone: ptr("+7 701 987 6543"), Telegram: ptr("@daulet_kz"),
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
			FullName: "Madina Orazova", Email: "madina.o@mail.com", Phone: ptr("+7 705 456 7890"), Telegram: ptr("@madina_or"),
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
			FullName: "Arman Tulegenov", Email: "arman.t@mail.com", Phone: ptr("+7 702 111 2233"), Telegram: ptr("@arman_t"),
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
			FullName: "Zhanna Bektursynova", Email: "zhanna.b@mail.com", Phone: ptr("+7 708 333 4455"), Telegram: ptr("@zhanna_bek"),
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
			FullName: "Nurlan Akhmetov", Email: "nurlan.a@mail.com", Phone: ptr("+7 700 555 6677"), Telegram: ptr("@nurlan_a"),
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
			FullName: "Kamila Sagynbaeva", Email: "kamila.s@mail.com", Phone: ptr("+7 706 777 8899"), Telegram: ptr("@kamila_sg"),
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
			FullName: "Bekzat Nurmagambetov", Email: "bekzat.n@mail.com", Phone: ptr("+7 771 222 3344"), Telegram: ptr("@bekzat_n"),
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
			FullName: "Dinara Ospanova", Email: "dinara.o@mail.com", Phone: ptr("+7 747 444 5566"), Telegram: ptr("@dinara_osp"),
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
			FullName: "Tamerlan Yessimov", Email: "tamerlan.y@mail.com", Phone: ptr("+7 778 666 7788"), Telegram: ptr("@tamerlan_y"),
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
			FullName: "Aisha Muratova", Email: "aisha.m@mail.com", Phone: ptr("+7 703 888 9900"), Telegram: ptr("@aisha_m"),
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
			FullName: "Yerbol Satybaldiev", Email: "yerbol.s@mail.com", Phone: ptr("+7 709 111 0011"), Telegram: ptr("@yerbol_sat"),
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
			FullName: "Saltanat Rakhimova", Email: "saltanat.r@mail.com", Phone: ptr("+7 775 222 3300"), Telegram: ptr("@saltanat_r"),
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
			FullName: "Rustem Torekulov", Email: "rustem.t@mail.com", Phone: ptr("+7 776 444 5500"), Telegram: ptr("@rustem_t"),
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
			FullName: "Daniyar Aubakirov", Email: "daniyar.a@mail.com", Phone: ptr("+7 704 555 6600"), Telegram: ptr("@daniyar_a"),
			Age: ptrInt(17), City: ptr("Astana"), School: ptr("School #55 Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Participated in hackathon (didn't win). Good grades in mathematics. Member of school robotics club."),
			Extracurriculars: ptr("Robotics club. Gaming club. Occasional volunteer."),
			Essay: `I want to study at inVision U because I believe education is the key to success. I have always been interested in technology and want to learn more about programming and entrepreneurship. My school has a robotics club where I learned basic Arduino programming. I also participated in a hackathon where our team tried to build a food delivery app.

I think inVision U is a great opportunity for me because of the scholarship and the quality of education. Kazakhstan needs more tech specialists and I want to be one of them. I am hardworking and motivated, and I believe I can succeed if given the chance.

I don't have many big achievements yet, but I am eager to learn and grow. I think university is where I will find my direction and start building meaningful projects.`,
			MotivationStatement: ptr("I want to become a successful entrepreneur and create technology that helps people. inVision U can give me the skills and network I need."),
		},
		{
			FullName: "Gulnaz Temirgaliyeva", Email: "gulnaz.t@mail.com", Phone: ptr("+7 707 666 7700"), Telegram: ptr("@gulnaz_tem"),
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
			FullName: "Miras Serikbayev", Email: "miras.s@mail.com", Phone: ptr("+7 701 777 8800"), Telegram: ptr("@miras_ser"),
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
			FullName: "Aliya Nurlankyzy", Email: "aliya.n@mail.com", Phone: ptr("+7 705 888 9900"), Telegram: ptr("@aliya_nurlan"),
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
			FullName: "Olzhas Kurmangaliyev", Email: "olzhas.k@mail.com", Phone: ptr("+7 708 999 0011"), Telegram: ptr("@olzhas_k"),
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
			FullName: "Zarina Yessenova", Email: "zarina.y@mail.com", Phone: ptr("+7 702 000 1122"), Telegram: ptr("@zarina_yes"),
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
			FullName: "Erlan Kazhimov", Email: "erlan.k@mail.com", Phone: ptr("+7 700 111 2200"), Telegram: ptr("@erlan_kazh"),
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
			FullName: "Inkar Baimukhanova", Email: "inkar.b@mail.com", Phone: ptr("+7 771 333 4400"), Telegram: ptr("@inkar_b"),
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
			FullName: "Askar Jumagulov", Email: "askar.j@mail.com", Phone: ptr("+7 747 555 6600"), Telegram: ptr("@askar_j"),
			Age: ptrInt(17), City: ptr("Almaty"), School: ptr("School #78 Almaty"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("None significant"),
			Extracurriculars: ptr("Gaming"),
			Essay: `I want to go to inVision U because it offers free education and I think it would be a good opportunity. I heard about it from a friend and decided to apply. I like technology and want to learn programming. I think I would be a good student because I am smart and learn quickly. Please give me a chance.`,
			MotivationStatement: ptr("I want free education and a good career."),
		},
		// AI-generated essay (HIGH risk)
		{
			FullName: "Sanzhar Omarov", Email: "sanzhar.o@mail.com", Phone: ptr("+7 778 666 7700"), Telegram: ptr("@sanzhar_o"),
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
			FullName: "Merey Iskakov", Email: "merey.i@mail.com", Phone: ptr("+7 706 777 8800"), Telegram: ptr("@merey_isk"),
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
			FullName: "Nurislam Bekzhanov", Email: "nurislam.b@mail.com", Phone: ptr("+7 700 888 9900"), Telegram: ptr("@nurislam_b"),
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
			FullName: "Ayaulym Serikova", Email: "ayaulym.s@mail.com", Phone: ptr("+7 701 999 0000"), Telegram: ptr("@ayaulym_s"),
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
			FullName: "Temirlan Zhumabekov", Email: "temirlan.z@mail.com", Phone: ptr("+7 775 000 1100"), Telegram: ptr("@temirlan_zh"),
			Age: ptrInt(18), City: ptr("Kyzylorda"), School: ptr("School #7 Kyzylorda"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Helped organize school concert. Basic English knowledge. Interested in business."),
			Extracurriculars: ptr("Music. Helping family with shop. Reading business books."),
			Essay: `My father has a small shop in Kyzylorda bazaar. I've been helping him since I was young. He taught me that business is about relationships — knowing your customers, being honest, and showing up every day even when you don't want to.

In today's digital economy, however, traditional business practices must evolve to remain competitive. The integration of technology into small business operations represents a critical pathway toward sustainable growth and enhanced market positioning. I believe that gaining expertise in digital transformation strategies would enable me to revolutionize my family's business model and contribute to the broader economic development of my region.

I want to learn at inVision U because I see how the world is changing and I don't want my city to be left behind. Kyzylorda needs people who understand both traditional business and new technology.`,
			MotivationStatement: ptr("I want to modernize business in Kyzylorda. My father's shop taught me the basics, but I need formal education to go further."),
		},
		{
			FullName: "Darkhan Moldabekov", Email: "darkhan.m@mail.com", Phone: ptr("+7 747 111 2200"), Telegram: ptr("@darkhan_m"),
			Age: ptrInt(18), City: ptr("Astana"), School: ptr("School #30 Astana"), GraduationYear: ptrInt(2026),
			Achievements:     ptr("Nothing notable"),
			Extracurriculars: ptr("Video games. Social media."),
			Essay: `I want to go to university for free thats why I am applying to inVision U. My grades are okay but not great. I like computers and games. I think I could learn programming if someone taught me properly. My school doesn't have good CS classes so I havent had the chance to learn much.

I live with my mom in Astana. She wants me to get higher education but we cant afford regular university. This scholarship is our best option.

I dont have achievements to list but I think potential matters more than past accomplishments. Everyone starts somewhere.`,
			MotivationStatement: ptr("Need scholarship for education. Will try hard."),
		},
	}

	// Generate additional candidates to reach 100
	candidates = append(candidates, generateAdditionalCandidates()...)

	inserted := 0
	failed := 0
	for _, c := range candidates {
		// Assign a random major if none specified
		major := c.Major
		if major == nil {
			tags := []string{"Engineering", "Tech", "Society", "Policy Reform", "Art + Media"}
			m := tags[rand.Intn(len(tags))]
			major = &m
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, major, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,'pending')`,
			c.FullName, c.Email, c.Phone, c.Telegram, c.Age, c.City, c.School, c.GraduationYear,
			c.Achievements, c.Extracurriculars, c.Essay, c.MotivationStatement, major,
		)
		if err != nil {
			log.Printf("SEED ERROR [%s / %s]: %v", c.FullName, c.Email, err)
			failed++
		} else {
			inserted++
		}
	}

	log.Printf("Seeding complete: %d inserted, %d failed (total attempted: %d)", inserted, failed, len(candidates))
	if inserted == 0 {
		return fmt.Errorf("all %d inserts failed — check logs above for details", len(candidates))
	}
	return nil
}

func generateAdditionalCandidates() []seedCandidate {
	r := rand.New(rand.NewSource(42)) // deterministic for reproducibility

	firstNames := []struct {
		name, gender string
	}{
		{"Adilkhan", "m"}, {"Nursultan", "m"}, {"Timur", "m"}, {"Samat", "m"}, {"Dias", "m"},
		{"Azamat", "m"}, {"Baurzhan", "m"}, {"Nurzhan", "m"}, {"Kanat", "m"}, {"Sultan", "m"},
		{"Zhandos", "m"}, {"Yernur", "m"}, {"Alibek", "m"}, {"Madi", "m"}, {"Ramazan", "m"},
		{"Aruzhan", "f"}, {"Togzhan", "f"}, {"Symbat", "f"}, {"Nazerke", "f"}, {"Aidana", "f"},
		{"Moldir", "f"}, {"Akbota", "f"}, {"Zhuldyz", "f"}, {"Karlygash", "f"}, {"Assem", "f"},
		{"Dana", "f"}, {"Aygerim", "f"}, {"Laura", "f"}, {"Amina", "f"}, {"Darina", "f"},
		{"Marlen", "m"}, {"Askhat", "m"}, {"Kuanysh", "m"}, {"Olzhas", "m"}, {"Yerzhan", "m"},
		{"Nurdaulet", "m"}, {"Madiyar", "m"}, {"Anelya", "f"}, {"Zhanerke", "f"}, {"Balnur", "f"},
	}

	lastNames := []string{
		"Abilov", "Tastanbekov", "Zhaksylykov", "Kairbekov", "Nurgaliyev",
		"Omarov", "Doszhanov", "Mukhamedjanov", "Satpayev", "Altynbekov",
		"Yermekova", "Dossanova", "Berkinbayeva", "Myrzakhmetova", "Tokanova",
		"Syzdykova", "Abdikerimova", "Nurpeissova", "Ibragimova", "Suleimenov",
		"Karimov", "Tulepov", "Zhumagaliyev", "Baitasov", "Sharipov",
		"Rakhmanova", "Ospanova", "Tastemirova", "Kasymova", "Amirova",
		"Zhunisov", "Kenzhebayev", "Abdullayev", "Smagulova", "Serikbayeva",
		"Kozhakhmetov", "Baytursynov", "Aitkalieva", "Nurmukanova", "Tulegenova",
	}

	cities := []struct {
		name, school string
	}{
		{"Almaty", "NIS PhM Almaty"}, {"Almaty", "Haileybury Almaty"}, {"Almaty", "KBTU Lyceum"},
		{"Almaty", "School #134 Almaty"}, {"Almaty", "Republican Physics-Math School"},
		{"Astana", "NIS Astana"}, {"Astana", "BIL Astana"}, {"Astana", "School #55 Astana"},
		{"Astana", "Miras International School"}, {"Shymkent", "NIS Shymkent"},
		{"Shymkent", "School #35 Shymkent"}, {"Aktobe", "NIS Aktobe"},
		{"Aktobe", "School #12 Aktobe"}, {"Karaganda", "NIS Karaganda"},
		{"Karaganda", "School #85 Karaganda"}, {"Atyrau", "NIS Atyrau"},
		{"Atyrau", "School #24 Atyrau"}, {"Pavlodar", "NIS Pavlodar"},
		{"Kostanay", "NIS Kostanay"}, {"Semey", "School #1 Semey"},
		{"Aktau", "School #14 Aktau"}, {"Taldykorgan", "School #5 Taldykorgan"},
		{"Oral", "School #3 Oral"}, {"Taraz", "NIS Taraz"},
		{"Kyzylorda", "School #7 Kyzylorda"}, {"Turkestan", "School #2 Turkestan"},
		{"Ekibastuz", "School #11 Ekibastuz"}, {"Temirtau", "School #6 Temirtau"},
	}

	type essayTemplate struct {
		category     string // "strong", "recommend", "borderline", "not_recommended"
		achievements string
		extras       string
		essay        string
		motivation   string
	}

	templates := []essayTemplate{
		// STRONG RECOMMEND templates
		{
			category:     "strong",
			achievements: "Founded a free tutoring program serving 150+ students across the region. National science olympiad gold medalist. Published research on renewable energy solutions in a youth science journal.",
			extras:       "President of school STEM club (2 years). Organizer of regional hackathon. Volunteer at local hospital.",
			essay: `When I was 15, I noticed something troubling: students in my city were failing science not because they lacked ability, but because they lacked access to quality instruction. Our school had one physics teacher for 400 students. Tutoring centers charged more than most families could afford.

So I started teaching. Every Saturday morning in the school library, then in community centers, then online. What began with 5 students grew to over 150 across multiple locations. I recruited top students as volunteer tutors, created a curriculum, and built a simple website to coordinate schedules.

The hardest part was convincing parents. Many didn't believe a teenager could teach effectively. I invited them to observe sessions. When they saw their children solving problems they'd struggled with for months, the skepticism disappeared.

My research on renewable energy emerged from a question one of my students asked: "Why doesn't %s use more solar power?" I didn't know the answer, so we investigated together. That investigation turned into a published paper analyzing solar potential in our region.

I want to join inVision U because I've learned to build educational programs from scratch, but I need to learn how to make them sustainable. I want to understand how inDrive built systems that scale — because I want to do the same for education access.`,
			motivation: "Education access is the defining challenge of our generation in Central Asia. I want inVision U to help me build solutions that last beyond my personal effort.",
		},
		{
			category:     "strong",
			achievements: "Created a mobile app for connecting local farmers with urban buyers (1000+ downloads). Winner of national innovation competition. Led a team of 8 developers in building school management software.",
			extras:       "Tech lead at school coding club. Mentor at regional startup incubator. Active open-source contributor.",
			essay: `My grandmother sells apples from her garden in a village 200km from %s. Every season, middlemen buy her entire harvest at a fraction of the market price. She has no choice — she has no way to reach city buyers directly.

This injustice drove me to build FarmConnect, an app that connects farmers directly with urban consumers. I taught myself React Native, built a backend in Node.js, and launched the app in three months. The first version was buggy and ugly, but it worked. My grandmother made 3x more selling through the app than through middlemen.

Now FarmConnect has over 1000 downloads and serves 50 farmers in our region. The technical challenges were significant — spotty rural internet, farmers who aren't tech-savvy, payment integration in a cash-heavy economy. Each obstacle taught me to design for real constraints, not ideal conditions.

Leading a team of 8 developers for our school management software was a different kind of challenge. I learned that the hardest part of building software isn't writing code — it's aligning people around a shared vision and making decisions when there's no clear right answer.

inVision U is where I want to be because I've proven I can build technology that helps people. Now I need to learn how to build a company around it.`,
			motivation: "Technology should serve the people who need it most, not just those who can afford it. inVision U's mission aligns with everything I'm building toward.",
		},
		// RECOMMEND templates
		{
			category:     "recommend",
			achievements: "Regional science olympiad medalist. Built a school library catalog system. Active volunteer at children's education center.",
			extras:       "Science club vice-president. Cross-country running. Community tutoring.",
			essay: `I've always been the person who notices inefficiencies. When our school librarian spent hours manually tracking books, I built a simple catalog system using Python and SQLite. When I saw younger kids struggling with math at the community center, I started tutoring twice a week.

These aren't revolutionary achievements. I know that. But each one taught me something. The library system taught me that even simple technology can save someone hours of tedious work. Tutoring taught me patience — when a 10-year-old doesn't understand fractions, you can't just explain it faster. You need to find a different way in.

The science olympiad pushed me academically in ways school classes don't. Competing against the best students in the region showed me how much I still have to learn. I placed well, but the students who beat me were operating at a level I want to reach.

I'm applying to inVision U because I want to be surrounded by people who are better than me — people who will push me to grow. In %s, I'm often the most technically capable person in the room. That's comfortable, but it's not how you improve.`,
			motivation: "I want to grow beyond what my current environment can offer. inVision U's community of driven students is what I need to reach my potential.",
		},
		{
			category:     "recommend",
			achievements: "Created a social media campaign for environmental awareness (3000+ reach). School debate champion. Completed online data science course.",
			extras:       "Debate team captain. Environmental club. Volunteer translator.",
			essay: `The environmental campaign I ran on social media started as a school project and turned into something bigger. I posted daily facts about pollution in Kazakhstan, created infographics about water scarcity, and organized an online pledge for reducing plastic use. Over 3,000 people engaged with the content.

What surprised me was how many people didn't know basic facts about environmental issues in their own country. The Aral Sea disaster, air pollution in major cities, soil degradation from industrial farming — these are well-documented problems, but most young people I spoke with knew very little about them.

Debate taught me to argue both sides of an issue. It's a skill that sounds adversarial but is actually about empathy — understanding why someone disagrees with you. My best debate performances came when I could genuinely see the strength in my opponent's position.

I've been learning data science because I believe data-driven decision making is the future of environmental policy. My online course taught me Python, pandas, and basic machine learning. I'm not an expert yet, but I can see how these tools could help analyze environmental data.

I want to study at inVision U because it combines technology, leadership, and real-world impact — the three things I'm most passionate about.`,
			motivation: "Environmental challenges need tech-savvy leaders. inVision U can help me become one.",
		},
		{
			category:     "recommend",
			achievements: "Organized a coding workshop for 60 students. School math team member. Interned at a local tech startup.",
			extras:       "Coding club founder. Track and field. Youth volunteer corps.",
			essay: `My summer internship at a small tech startup in %s was eye-opening. I went in expecting to write code all day. Instead, I spent most of my time in meetings, reviewing user feedback, and helping with customer support. The founder told me: "We don't build what we want to build. We build what users need."

That lesson shaped how I approach everything now. When I organized a coding workshop at school, I didn't teach what I thought was cool — I surveyed students first. They wanted to learn how to build personal websites and simple games, not algorithms. So that's what we taught. 60 students attended, and many continued coding afterward.

Being on the math team taught me discipline. You can't cram for math competitions — understanding builds slowly through consistent practice. I apply the same principle to coding: a little bit every day compounds into real skill.

I want to attend inVision U because the startup experience showed me what's possible, and now I want the skills to build something of my own.`,
			motivation: "I've seen how startups work from the inside, and I want to learn how to build one that solves real problems.",
		},
		{
			category:     "recommend",
			achievements: "Won school innovation fair. Built a Telegram bot for student announcements. Peer tutor in physics and math.",
			extras:       "Innovation club. Badminton team. School newspaper contributor.",
			essay: `The Telegram bot I built for our school sends daily announcements, schedule changes, and exam reminders to 300 students. Before the bot, information spread through word of mouth, which meant half the school missed important updates. Now everyone gets the same information at the same time.

Building it was straightforward technically — Python with the Telegram API. The hard part was adoption. Students installed it but didn't check notifications. I added a daily quiz feature with small prizes, and engagement tripled. This taught me that useful isn't enough — you also need to be engaging.

The innovation fair project was a prototype for a smart waste bin that sorts recyclables using image recognition. We trained a basic model on photos of common waste items. The accuracy was only 70%, but the concept won because the judges saw practical potential.

Tutoring physics and math gives me joy because it forces me to truly understand concepts. You can't explain something clearly if your own understanding is fuzzy. Every student I tutor makes me a better learner.

inVision U attracts the kind of people I want to learn alongside — builders who care about impact.`,
			motivation: "I build tools that solve everyday problems. inVision U will help me think bigger and build better.",
		},
		// BORDERLINE templates
		{
			category:     "borderline",
			achievements: "Participated in school science fair. Learning to code online. Average academic standing.",
			extras:       "Football team. Occasional volunteering. Interested in technology.",
			essay: `I'm not going to pretend I have a long list of achievements. I don't. What I have is curiosity and determination. I've been learning to code through free online courses for the past six months. Progress is slow, but I'm sticking with it.

My science fair project was a simple weather display using an Arduino. It didn't win anything, but building it felt amazing. Taking something from an idea to a working device — even a simple one — showed me what's possible.

I come from a family where nobody went to university. My parents work hard to provide for us, and they support my dream of getting higher education, even though they don't fully understand what computer science is.

I want inVision U to be the place where I figure out what I'm capable of. I know I'm not the strongest applicant, but I believe in my ability to work hard and grow. Sometimes all you need is the right environment and a chance.`,
			motivation: "I need an opportunity to prove myself. My background doesn't reflect my potential.",
		},
		{
			category:     "borderline",
			achievements: "School English competition participant. Helped organize charity event. Self-studying web development.",
			extras:       "English club. Some charity work. Self-study.",
			essay: `I've been interested in technology for as long as I can remember, but opportunities in %s are limited. There are no coding bootcamps, no tech meetups, no mentors. I learn everything from the internet.

Last year, I helped organize a charity event for a local orphanage. We collected books, clothes, and school supplies. It wasn't a huge event, but seeing the children's faces when they received new backpacks made all the organizing worthwhile.

In the English competition, I made it to the city finals but didn't place. The experience was valuable though — it showed me that my English is good enough to compete but needs improvement. I've been watching English YouTube channels daily to get better.

I'm self-studying web development and can build basic websites. It's frustrating sometimes — when you get stuck on a bug with no one to ask, it can take days to solve what a mentor could explain in minutes.

I'm applying to inVision U because I want access to the mentorship and community that I've been missing.`,
			motivation: "Limited resources shouldn't limit dreams. I want inVision U to give me the tools I lack.",
		},
		{
			category:     "borderline",
			achievements: "Good grades in STEM subjects. Created a blog about technology trends. Participated in online hackathon.",
			extras:       "Tech blog. Online coding communities. Reading.",
			essay: `I write a blog about technology trends in Kazakhstan. It has about 200 regular readers — mostly my classmates and their friends. I cover topics like how AI is changing different industries, new startups in Central Asia, and tech career paths.

Writing the blog forces me to research and understand topics deeply. When I wrote about blockchain, I went down a rabbit hole that lasted two weeks. I emerged with a much clearer understanding of the technology and its potential applications in our region.

I participated in an online hackathon where our team built a concept for a student mental health app. We didn't finish the prototype in time, but the experience of working with strangers from different countries under a tight deadline was valuable.

My grades in STEM subjects are good, but I know grades alone don't make someone a good candidate. I want to do more — build things, lead projects, make an impact. I just haven't had the right platform yet.

inVision U could be that platform for me.`,
			motivation: "I understand technology but haven't yet applied it to real problems. inVision U can bridge that gap.",
		},
		// NOT RECOMMENDED templates
		{
			category:     "not_recommended",
			achievements: "No significant achievements.",
			extras:       "Social media. Gaming.",
			essay: `I want to study at inVision U because it is free and has good education. I like technology and computers. I spend a lot of time online and I think I could be good at programming if I learned it properly. I dont have many achievements but I think university is where you start achieving things not before. Please consider my application.`,
			motivation: "Free education and good career opportunities.",
		},
		{
			category:     "not_recommended",
			achievements: "Participated in school event. Below average grades.",
			extras:       "Social media. Hanging out with friends.",
			essay: `In an era of unprecedented technological advancement, the imperative for quality education has never been more pronounced. I am writing to express my fervent desire to join the esteemed inVision U program, which I believe represents the apogee of educational excellence in the Central Asian region.

My journey, though perhaps not adorned with conventional accolades, has been characterized by an unwavering commitment to personal growth and a profound appreciation for the transformative power of knowledge. I am confident that the synergistic interplay between my innate curiosity and inVision U's world-class curriculum will yield exceptional outcomes.

I eagerly anticipate the opportunity to contribute to and benefit from the dynamic intellectual ecosystem at inVision U.`,
			motivation: "I am driven by an insatiable thirst for knowledge and a deep commitment to leveraging technology for societal betterment.",
		},
		{
			category:     "not_recommended",
			achievements: "Nothing specific.",
			extras:       "TikTok content creation. Gaming tournaments.",
			essay: `im applying because my friend told me about this university and the scholarship sounds really good. i dont really know what i want to study but technology seems interesting. i make tiktok videos sometimes and some of them got a few thousand views so i think i understand social media and digital stuff.

i know i should probably write more about my achievements but honestly i havent done much yet. high school is boring and i cant wait to be done with it. university will be different i think.

give me a chance and i wont disappoint you.`,
			motivation: "want to try something new and the scholarship is a good deal",
		},
	}

	var additional []seedCandidate
	usedNames := make(map[string]bool)
	idx := 0

	// Target distribution: ~10 strong, ~25 recommend, ~20 borderline, ~16 not_recommended = 71
	distribution := []struct {
		category string
		count    int
	}{
		{"strong", 10},
		{"recommend", 25},
		{"borderline", 20},
		{"not_recommended", 16},
	}

	for _, dist := range distribution {
		// Collect templates for this category
		var catTemplates []essayTemplate
		for _, t := range templates {
			if t.category == dist.category {
				catTemplates = append(catTemplates, t)
			}
		}

		for i := 0; i < dist.count; i++ {
			// Pick a unique name
			var fullName, email string
			for {
				fn := firstNames[r.Intn(len(firstNames))]
				ln := lastNames[r.Intn(len(lastNames))]
				fullName = fn.name + " " + ln
				if !usedNames[fullName] {
					usedNames[fullName] = true
					email = fmt.Sprintf("%s.%s.%d@mail.com",
						strings.ToLower(fn.name), strings.ToLower(ln)[:3], idx+30)
					break
				}
			}

			city := cities[r.Intn(len(cities))]
			tmpl := catTemplates[i%len(catTemplates)]
			age := 17 + r.Intn(2)

			essay := fmt.Sprintf(tmpl.essay, city.name)
			// Remove excess %s if template doesn't use city
			if !strings.Contains(tmpl.essay, "%s") {
				essay = tmpl.essay
			}

			phone := fmt.Sprintf("+7 7%02d %03d %04d", r.Intn(100), r.Intn(1000), r.Intn(10000))
			nameParts := strings.SplitN(strings.ToLower(fullName), " ", 2)
			tgUser := fmt.Sprintf("@%s_%s_%d", nameParts[0], nameParts[1][:3], idx+30)

			additional = append(additional, seedCandidate{
				FullName:            fullName,
				Email:               email,
				Phone:               ptr(phone),
				Telegram:            ptr(tgUser),
				Age:                 ptrInt(age),
				City:                ptr(city.name),
				School:              ptr(city.school),
				GraduationYear:      ptrInt(2026),
				Achievements:        ptr(tmpl.achievements),
				Extracurriculars:    ptr(tmpl.extras),
				Essay:               essay,
				MotivationStatement: ptr(tmpl.motivation),
			})
			idx++
		}
	}

	return additional
}
