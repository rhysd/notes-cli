#!/usr/bin/env ruby

require 'pathname'
require 'fileutils'
require 'date'
require 'mkmf'

def help
  puts "#{$PROGRAM_NAME} <memolist dir> <notes-cli home>"
end

Memo = Struct.new(:file, :title, :date, :category, :tags, :body)
TIMEZONE = Time.now.zone

def fail_read(path)
  raise "Invalid memo at #{path}"
end

def read_memo(path)
  puts "Reading memo #{path}"
  File.basename(path) =~ /^\d+-\d+-\d+-(.*\.md)$/
  file = $1
  lines = File.readlines path
  title = lines.shift.sub(/^title: /, '').chop
  fail_read path unless lines.shift =~ /^=+$/
  date = DateTime.parse "#{lines.shift.sub(/^date: /, '').chop} #{TIMEZONE}"
  tags = lines.shift.gsub(/^tags: \[|\]$/, '').chop.split(',').reject{|s| s == "" }
  categories = lines.shift.gsub(/^categories: \[|\]$/, '').chop.split(',').reject{|s| s == "" }
  fail_read path unless lines.shift =~ /^\s{0,3}(?:-+\s*){3,}$/
  body = lines.join.strip
  if categories.length == 1
    category = categories.first
  else
    category = "imported"
    tags += categories
  end
  Memo.new(file, title, date, category, tags, body)
end

def migrate(memo, home_dir)
  dir = home_dir.join(memo.category)
  FileUtils.mkdir_p dir.to_s unless dir.exist?
  file = dir.join(memo.file).to_s
  puts "Generating note #{file}"
  File.open(file, 'w') do |f|
    f.write <<~EOS
    #{memo.title}
    #{'=' * memo.title.length}
    - Category: #{memo.category}
    - Tags: #{memo.tags.join(", ")}
    - Created: #{memo.date.rfc3339}
    ---

    #{memo.body}
    EOS
  end
end

def main
  if ARGV.length < 2
    help
    exit 1
  end

  memolist_dir = Pathname.new ARGV[0]
  notes_home = Pathname.new(ARGV[1])

  puts "Migrating from memolist '#{memolist_dir}' to notes-cli '#{notes_home}'"

  unless memolist_dir.exist?
    puts "memolist dir not exist"
    exit 1
  end

  FileUtils.mkdir_p notes_home.to_s unless notes_home.exist?

  Dir.glob("#{memolist_dir}/*.md").map{|p| read_memo p}.each do |path|
    migrate(path, notes_home)
  end

  git = find_executable('git')
  if git
    system(git, '-C', notes_home.to_s, 'init')
  end
end

if __FILE__ == $PROGRAM_NAME
  main
end
