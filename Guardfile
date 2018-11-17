YELLOW='[93m'
RESET='[0m'

def run(cmdline)
  puts "#{YELLOW}+#{cmdline}#{RESET}"
  system cmdline
end

guard :shell do
  watch /\.go$/ do |m|
    puts "#{Time.now}: #{m[0]}"
    case m[0]
    when /_test\.go$/
      parent = File.dirname m[0]
      sources = Dir["#{parent}/*.go"].reject{|p| %w(_test.go _other.go).any?{|s| p.end_with? s } }
      sources << m[0] << "common_test.go"
      # Assume that https://github.com/rhysd/gotest is installed
      run "gotest #{sources.uniq.join ' '}"
      run "golint #{m[0]}"
    else
      run 'go build ./cmd/notes'
      run "golint #{m[0]}"
    end
  end
end
