use egg_mode::error::Result;

#[tokio::main]
async fn main() -> Result<()> {
    let key = include_str!("credentials/key").trim();
    let secret = include_str!("credentials/secret").trim();

    let con_token = egg_mode::KeyPair::new(key, secret);

    println!("Pulling up the bearer token...");
    let token = egg_mode::auth::bearer_token(&con_token).await?;

    println!("Pulling up a user timeline...");
    let timeline =
        egg_mode::tweet::user_timeline("rustlang", false, true, &token).with_page_size(5);

    let (_timeline, feed) = timeline.start().await?;
    for tweet in feed.response {
        println!("");
        common::print_tweet(&tweet);
    }
    Ok(())
}
