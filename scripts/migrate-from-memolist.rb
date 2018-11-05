#!/usr/bin/env ruby

require 'pathname'
require 'fileutils'
require 'date'

def help
  puts "#{$PROGRAM_NAME} <memolist dir> <notes-cli home>"
end

Memo = Struct.new(:file, :title, :date, :tags, :body)
TIMEZONE = Time.now.zone

def fail_read(path)
  raise "Invalid memo at #{path}"
end

def read_memo(path)
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
  Memo.new(file, title, date, tags + categories, body)
end

def migrate(memo, dest_dir)
  File.open(dest_dir.join(memo.file).to_s, 'w') do |f|
    f.write <<~EOS
    #{memo.title}
    #{'=' * memo.title.length}
    - Category: imported
    - Tags: #{memo.tags}
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
  import_dir = Pathname.new(ARGV[1]).join('imported')

  puts "Migrating from memolist '#{memolist_dir}' to notes-cli '#{import_dir}'"

  unless memolist_dir.exist?
    puts "memolist dir not exist"
    exit 1
  end

  FileUtils.mkdir_p import_dir.to_s unless import_dir.exist?

  Dir.glob("#{memolist_dir}/*.md").map{|p| read_memo p}.each do |path|
    migrate(path, import_dir)
  end
end

#
# main
#
if __FILE__ == $PROGRAM_NAME
  main
end
